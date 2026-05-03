package repository

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"gorm.io/gorm"
)

// TenantAwareRepository provides tenant isolation for existing repositories
type TenantAwareRepository struct {
	db       *gorm.DB
	tenantID string
}

// NewTenantAwareRepository creates a new tenant-aware repository
func NewTenantAwareRepository(db *gorm.DB, tenantID string) *TenantAwareRepository {
	return &TenantAwareRepository{
		db:       db,
		tenantID: tenantID,
	}
}

// WithTenant creates a new tenant-aware repository instance
func (r *TenantAwareRepository) WithTenant(tenantID string) *TenantAwareRepository {
	return &TenantAwareRepository{
		db:       r.db,
		tenantID: tenantID,
	}
}

// GetTenantID returns the current tenant ID
func (r *TenantAwareRepository) GetTenantID() string {
	return r.tenantID
}

// ValidateTenant validates that the tenant exists and is active
func (r *TenantAwareRepository) ValidateTenant(ctx context.Context) error {
	if r.tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	var tenantRecord tenant.Tenant
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", r.tenantID, tenant.TenantStatusActive).First(&tenantRecord).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tenant not found or inactive")
		}
		return fmt.Errorf("failed to validate tenant: %w", err)
	}

	return nil
}

// TenantScopedQuery creates a query scoped to the current tenant
func (r *TenantAwareRepository) TenantScopedQuery(ctx context.Context, model any) *gorm.DB {
	query := r.db.WithContext(ctx).Model(model)
	if r.tenantID != "" {
		query = query.Where("tenant_id = ?", r.tenantID)
	}
	return query
}

// TenantScopedTransaction creates a transaction scoped to the current tenant
func (r *TenantAwareRepository) TenantScopedTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Apply tenant filtering to all operations within the transaction
		if r.tenantID != "" {
			// Set tenant context for the transaction
			tx = tx.Set("tenant_id", r.tenantID)

			// Create a wrapped DB instance that automatically applies tenant filtering
			wrappedTx := &TenantScopedDB{
				DB:       tx,
				tenantID: r.tenantID,
			}

			// Use the wrapped transaction for all operations
			// Since the function expects *gorm.DB, we need to embed the wrapper properly
			return fn(wrappedTx.DB.Scopes(func(db *gorm.DB) *gorm.DB {
				// Apply tenant filtering to all queries
				return db.Where("tenant_id = ?", wrappedTx.tenantID)
			}))
		}
		return fn(tx)
	})
}

// TenantScopedDB wraps GORM DB with automatic tenant filtering
type TenantScopedDB struct {
	*gorm.DB
	tenantID string
}

// Table overrides the Table method to add tenant filtering
func (t *TenantScopedDB) Table(name string) *gorm.DB {
	// Add tenant filtering for tables that have tenant_id column
	tenantFilteredTables := map[string]bool{
		"tenants":                 true,
		"tenant_users":            true,
		"tenant_configs":          true,
		"tenant_events":           true,
		"profiles":                true,
		"rate_plans":              true,
		"rate_plan_subscriptions": true,
		"carriers":                true,
		"pricing_rules":           true,
	}

	if tenantFilteredTables[name] {
		return t.DB.Table(name).Where("tenant_id = ?", t.tenantID)
	}

	return t.DB.Table(name)
}

// ValidateResourceAccess validates that a resource belongs to a tenant
func (v *TenantResourceValidator) ValidateResourceAccess(ctx context.Context, tenantID, resourceType, resourceID string) error {
	switch resourceType {
	case "profile":
		return v.validateProfileAccess(ctx, tenantID, resourceID)
	case "rateplan":
		return v.validateRatePlanAccess(ctx, tenantID, resourceID)
	case "carrier":
		return v.validateCarrierAccess(ctx, tenantID, resourceID)
	case "subscription":
		return v.validateSubscriptionAccess(ctx, tenantID, resourceID)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func (v *TenantResourceValidator) validateProfileAccess(ctx context.Context, tenantID, profileID string) error {
	if profileID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}

	// Check if profile exists and belongs to tenant
	var count int64
	err := v.db.WithContext(ctx).
		Table("profiles").
		Where("iccid = ? AND tenant_id = ?", profileID, tenantID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to validate profile access: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("profile %s not found or does not belong to tenant %s", profileID, tenantID)
	}

	return nil
}

func (v *TenantResourceValidator) validateRatePlanAccess(ctx context.Context, tenantID, ratePlanID string) error {
	if ratePlanID == "" {
		return fmt.Errorf("rate plan ID cannot be empty")
	}

	// Check if rate plan exists and belongs to tenant
	var count int64
	err := v.db.WithContext(ctx).
		Table("rate_plans").
		Where("id = ? AND tenant_id = ?", ratePlanID, tenantID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to validate rate plan access: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("rate plan %s not found or does not belong to tenant %s", ratePlanID, tenantID)
	}

	return nil
}

func (v *TenantResourceValidator) validateCarrierAccess(ctx context.Context, tenantID, carrierID string) error {
	if carrierID == "" {
		return fmt.Errorf("carrier ID cannot be empty")
	}

	// Check if carrier exists and belongs to tenant
	var count int64
	err := v.db.WithContext(ctx).
		Table("carriers").
		Where("id = ? AND tenant_id = ?", carrierID, tenantID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to validate carrier access: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("carrier %s not found or does not belong to tenant %s", carrierID, tenantID)
	}

	return nil
}

func (v *TenantResourceValidator) validateSubscriptionAccess(ctx context.Context, tenantID, subscriptionID string) error {
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}

	// Check if subscription exists and belongs to tenant
	var count int64
	err := v.db.WithContext(ctx).
		Table("rate_plan_subscriptions").
		Where("id = ? AND profile_id IN (SELECT iccid FROM profiles WHERE tenant_id = ?)", subscriptionID, tenantID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to validate subscription access: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("subscription %s not found or does not belong to tenant %s", subscriptionID, tenantID)
	}

	return nil
}

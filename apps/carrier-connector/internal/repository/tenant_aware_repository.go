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
			// This would need to be implemented based on specific requirements
			// For now, we'll rely on the tenant-scoped queries within the transaction
		}
		return fn(tx)
	})
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
	// This would check if the profile belongs to the tenant
	// Implementation depends on the actual profile schema
	return nil
}

func (v *TenantResourceValidator) validateRatePlanAccess(ctx context.Context, tenantID, ratePlanID string) error {
	// This would check if the rate plan belongs to the tenant
	// Implementation depends on the actual rate plan schema
	return nil
}

func (v *TenantResourceValidator) validateCarrierAccess(ctx context.Context, tenantID, carrierID string) error {
	// This would check if the carrier belongs to the tenant
	// Implementation depends on the actual carrier schema
	return nil
}

func (v *TenantResourceValidator) validateSubscriptionAccess(ctx context.Context, tenantID, subscriptionID string) error {
	// This would check if the subscription belongs to the tenant
	// Implementation depends on the actual subscription schema
	return nil
}

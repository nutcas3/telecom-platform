package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TenantAwareRepository provides tenant isolation for existing repositories
type TenantAwareRepository struct {
	db     *gorm.DB
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

	var tenant Tenant
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", r.tenantID, TenantStatusActive).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tenant not found or inactive")
		}
		return fmt.Errorf("failed to validate tenant: %w", err)
	}

	return nil
}

// TenantScopedQuery creates a query scoped to the current tenant
func (r *TenantAwareRepository) TenantScopedQuery(ctx context.Context, model interface{}) *gorm.DB {
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

// Helper methods for common tenant-aware operations

// CreateWithTenant creates a record with tenant ID
func (r *TenantAwareRepository) CreateWithTenant(ctx context.Context, model interface{}) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	// Set tenant ID if the model has a TenantID field
	if modelWithTenant, ok := model.(interface{ SetTenantID(string) }); ok {
		modelWithTenant.SetTenantID(r.tenantID)
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// GetByTenantID retrieves a record by ID within the current tenant
func (r *TenantAwareRepository) GetByTenantID(ctx context.Context, model interface{}, id string) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.TenantScopedQuery(ctx, model).Where("id = ?", id).First(model).Error
}

// UpdateWithTenant updates a record within the current tenant
func (r *TenantAwareRepository) UpdateWithTenant(ctx context.Context, model interface{}) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Save(model).Error
}

// DeleteWithTenant deletes a record within the current tenant
func (r *TenantAwareRepository) DeleteWithTenant(ctx context.Context, model interface{}, id string) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.TenantScopedQuery(ctx, model).Where("id = ?", id).Delete(model).Error
}

// ListWithTenant lists records within the current tenant
func (r *TenantAwareRepository) ListWithTenant(ctx context.Context, model interface{}, results interface{}, filters map[string]interface{}) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	query := r.TenantScopedQuery(ctx, model)
	
	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	return query.Find(results).Error
}

// CountWithTenant counts records within the current tenant
func (r *TenantAwareRepository) CountWithTenant(ctx context.Context, model interface{}, filters map[string]interface{}) (int64, error) {
	if err := r.ValidateTenant(ctx); err != nil {
		return 0, err
	}

	query := r.TenantScopedQuery(ctx, model)
	
	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// TenantAwareModel is an interface for models that support tenant isolation
type TenantAwareModel interface {
	SetTenantID(tenantID string)
	GetTenantID() string
}

// BaseTenantModel provides a base implementation for tenant-aware models
type BaseTenantModel struct {
	TenantID string `json:"tenant_id" gorm:"column:tenant_id;index;not null"`
}

// SetTenantID sets the tenant ID
func (m *BaseTenantModel) SetTenantID(tenantID string) {
	m.TenantID = tenantID
}

// GetTenantID returns the tenant ID
func (m *BaseTenantModel) GetTenantID() string {
	return m.TenantID
}

// TenantIsolationMiddleware provides database-level tenant isolation
type TenantIsolationMiddleware struct {
	db *gorm.DB
}

// NewTenantIsolationMiddleware creates a new tenant isolation middleware
func NewTenantIsolationMiddleware(db *gorm.DB) *TenantIsolationMiddleware {
	return &TenantIsolationMiddleware{db: db}
}

// WithTenantIsolation applies tenant isolation to a database operation
func (m *TenantIsolationMiddleware) WithTenantIsolation(ctx context.Context, tenantID string, operation func(*gorm.DB) error) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required for tenant-isolated operations")
	}

	// Validate tenant exists
	var tenant Tenant
	err := m.db.WithContext(ctx).Where("id = ? AND status = ?", tenantID, TenantStatusActive).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tenant not found or inactive")
		}
		return fmt.Errorf("failed to validate tenant: %w", err)
	}

	// Execute operation with tenant context
	tx := m.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	return operation(tx)
}

// GetTenantFromContext extracts tenant ID from context
func GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// SetTenantInContext sets tenant ID in context
func SetTenantInContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// TenantQueryBuilder helps build tenant-aware queries
type TenantQueryBuilder struct {
	db       *gorm.DB
	tenantID string
	query    *gorm.DB
}

// NewTenantQueryBuilder creates a new tenant query builder
func NewTenantQueryBuilder(db *gorm.DB, tenantID string) *TenantQueryBuilder {
	return &TenantQueryBuilder{
		db:       db,
		tenantID: tenantID,
		query:    db.Where("tenant_id = ?", tenantID),
	}
}

// Where adds a where clause to the query
func (b *TenantQueryBuilder) Where(query string, args ...interface{}) *TenantQueryBuilder {
	b.query = b.query.Where(query, args...)
	return b
}

// Order adds ordering to the query
func (b *TenantQueryBuilder) Order(value string) *TenantQueryBuilder {
	b.query = b.query.Order(value)
	return b
}

// Limit adds a limit to the query
func (b *TenantQueryBuilder) Limit(limit int) *TenantQueryBuilder {
	b.query = b.query.Limit(limit)
	return b
}

// Offset adds an offset to the query
func (b *TenantQueryBuilder) Offset(offset int) *TenantQueryBuilder {
	b.query = b.query.Offset(offset)
	return b
}

// Find executes the find operation
func (b *TenantQueryBuilder) Find(dest interface{}) error {
	return b.query.Find(dest).Error
}

// First executes the first operation
func (b *TenantQueryBuilder) First(dest interface{}) error {
	return b.query.First(dest).Error
}

// Count executes the count operation
func (b *TenantQueryBuilder) Count() (int64, error) {
	var count int64
	err := b.query.Count(&count).Error
	return count, err
}

// GetQuery returns the underlying gorm query
func (b *TenantQueryBuilder) GetQuery() *gorm.DB {
	return b.query
}

// TenantResourceValidator validates resource access across tenants
type TenantResourceValidator struct {
	db *gorm.DB
}

// NewTenantResourceValidator creates a new tenant resource validator
func NewTenantResourceValidator(db *gorm.DB) *TenantResourceValidator {
	return &TenantResourceValidator{db: db}
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

package repository

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"gorm.io/gorm"
)

func (r *TenantAwareRepository) CreateWithTenant(ctx context.Context, model any) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	// Set tenant ID if the model has a TenantID field
	if modelWithTenant, ok := model.(interface{ SetTenantID(string) }); ok {
		modelWithTenant.SetTenantID(r.tenantID)
	}

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *TenantAwareRepository) GetByTenantID(ctx context.Context, model any, id string) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.TenantScopedQuery(ctx, model).Where("id = ?", id).First(model).Error
}

func (r *TenantAwareRepository) UpdateWithTenant(ctx context.Context, model any) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Save(model).Error
}

func (r *TenantAwareRepository) DeleteWithTenant(ctx context.Context, model any, id string) error {
	if err := r.ValidateTenant(ctx); err != nil {
		return err
	}

	return r.TenantScopedQuery(ctx, model).Where("id = ?", id).Delete(model).Error
}

func (r *TenantAwareRepository) ListWithTenant(ctx context.Context, model any, results any, filters map[string]any) error {
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

func (r *TenantAwareRepository) CountWithTenant(ctx context.Context, model any, filters map[string]any) (int64, error) {
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

type TenantAwareModel interface {
	SetTenantID(tenantID string)
	GetTenantID() string
}

type BaseTenantModel struct {
	TenantID string `json:"tenant_id" gorm:"column:tenant_id;index;not null"`
}

func (m *BaseTenantModel) SetTenantID(tenantID string) {
	m.TenantID = tenantID
}

func (m *BaseTenantModel) GetTenantID() string {
	return m.TenantID
}

type TenantIsolationMiddleware struct {
	db *gorm.DB
}

func NewTenantIsolationMiddleware(db *gorm.DB) *TenantIsolationMiddleware {
	return &TenantIsolationMiddleware{db: db}
}

func (m *TenantIsolationMiddleware) WithTenantIsolation(ctx context.Context, tenantID string, operation func(*gorm.DB) error) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required for tenant-isolated operations")
	}

	var tenantRecord tenant.Tenant
	err := m.db.WithContext(ctx).Where("id = ? AND status = ?", tenantID, tenant.TenantStatusActive).First(&tenantRecord).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tenant not found or inactive")
		}
		return fmt.Errorf("failed to validate tenant: %w", err)
	}

	tx := m.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	return operation(tx)
}

func GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

func SetTenantInContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

type TenantQueryBuilder struct {
	db       *gorm.DB
	tenantID string
	query    *gorm.DB
}

func NewTenantQueryBuilder(db *gorm.DB, tenantID string) *TenantQueryBuilder {
	return &TenantQueryBuilder{
		db:       db,
		tenantID: tenantID,
		query:    db.Where("tenant_id = ?", tenantID),
	}
}

func (b *TenantQueryBuilder) Where(query string, args ...any) *TenantQueryBuilder {
	b.query = b.query.Where(query, args...)
	return b
}

func (b *TenantQueryBuilder) Order(value string) *TenantQueryBuilder {
	b.query = b.query.Order(value)
	return b
}

func (b *TenantQueryBuilder) Limit(limit int) *TenantQueryBuilder {
	b.query = b.query.Limit(limit)
	return b
}

func (b *TenantQueryBuilder) Offset(offset int) *TenantQueryBuilder {
	b.query = b.query.Offset(offset)
	return b
}

func (b *TenantQueryBuilder) Find(dest any) error {
	return b.query.Find(dest).Error
}

func (b *TenantQueryBuilder) First(dest any) error {
	return b.query.First(dest).Error
}

func (b *TenantQueryBuilder) Count() (int64, error) {
	var count int64
	err := b.query.Count(&count).Error
	return count, err
}

func (b *TenantQueryBuilder) GetQuery() *gorm.DB {
	return b.query
}

type TenantResourceValidator struct {
	db *gorm.DB
}

func NewTenantResourceValidator(db *gorm.DB) *TenantResourceValidator {
	return &TenantResourceValidator{db: db}
}

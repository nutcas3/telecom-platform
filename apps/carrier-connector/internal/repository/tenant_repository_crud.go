package repository

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"gorm.io/gorm"
)

type GormTenantRepository struct {
	db *gorm.DB
}

func NewGormTenantRepository(db *gorm.DB) TenantRepository {
	return &GormTenantRepository{db: db}
}

func (r *GormTenantRepository) CreateTenant(ctx context.Context, tenant *tenant.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *GormTenantRepository) GetTenant(ctx context.Context, id string) (*tenant.Tenant, error) {
	var tenant tenant.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (r *GormTenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*tenant.Tenant, error) {
	var tenant tenant.Tenant
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// UpdateTenant updates an existing tenant
func (r *GormTenantRepository) UpdateTenant(ctx context.Context, tenant *tenant.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// DeleteTenant deletes a tenant
func (r *GormTenantRepository) DeleteTenant(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&tenant.Tenant{}, "id = ?", id).Error
}

// ListTenants lists tenants with filtering
func (r *GormTenantRepository) ListTenants(ctx context.Context, filter *tenant.TenantFilter) ([]*tenant.Tenant, error) {
	query := r.db.WithContext(ctx).Model(&tenant.Tenant{})

	if filter.ID != "" {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Domain != "" {
		query = query.Where("domain ILIKE ?", "%"+filter.Domain+"%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Plan != "" {
		query = query.Where("plan = ?", filter.Plan)
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var tenants []*tenant.Tenant
	err := query.Find(&tenants).Error
	return tenants, err
}

func (r *GormTenantRepository) CountTenants(ctx context.Context, filter *tenant.TenantFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&tenant.Tenant{})

	if filter.ID != "" {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Domain != "" {
		query = query.Where("domain ILIKE ?", "%"+filter.Domain+"%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Plan != "" {
		query = query.Where("plan = ?", filter.Plan)
	}

	var count int64
	err := query.Count(&count).Error
	return int(count), err
}

func (r *GormTenantRepository) CreateTenantUser(ctx context.Context, user *tenant.TenantUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormTenantRepository) GetTenantUser(ctx context.Context, tenantID, userID string) (*tenant.TenantUser, error) {
	var user tenant.TenantUser
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormTenantRepository) UpdateTenantUser(ctx context.Context, user *tenant.TenantUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormTenantRepository) DeleteTenantUser(ctx context.Context, tenantID, userID string) error {
	return r.db.WithContext(ctx).Delete(&tenant.TenantUser{}, "tenant_id = ? AND user_id = ?", tenantID, userID).Error
}

func (r *GormTenantRepository) ListTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) ([]*tenant.TenantUser, error) {
	query := r.db.WithContext(ctx).Model(&tenant.TenantUser{})

	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filter.Email+"%")
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var users []*tenant.TenantUser
	err := query.Find(&users).Error
	return users, err
}

func (r *GormTenantRepository) CountTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&tenant.TenantUser{})

	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filter.Email+"%")
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var count int64
	err := query.Count(&count).Error
	return int(count), err
}

package repository

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// CreateAPIKey creates a new API key
func (r *GormTenantRepository) CreateAPIKey(ctx context.Context, apiKey *tenant.TenantAPIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetAPIKey retrieves an API key by ID
func (r *GormTenantRepository) GetAPIKey(ctx context.Context, id string) (*tenant.TenantAPIKey, error) {
	var apiKey tenant.TenantAPIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetAPIKeyByHash retrieves an API key by hash
func (r *GormTenantRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*tenant.TenantAPIKey, error) {
	var apiKey tenant.TenantAPIKey
	err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateAPIKey updates an API key
func (r *GormTenantRepository) UpdateAPIKey(ctx context.Context, apiKey *tenant.TenantAPIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// DeleteAPIKey deletes an API key
func (r *GormTenantRepository) DeleteAPIKey(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&tenant.TenantAPIKey{}, "id = ?", id).Error
}

// ListAPIKeys lists API keys for a tenant
func (r *GormTenantRepository) ListAPIKeys(ctx context.Context, tenantID string) ([]*tenant.TenantAPIKey, error) {
	var apiKeys []*tenant.TenantAPIKey
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&apiKeys).Error
	return apiKeys, err
}

// CreateUsage creates a new usage record
func (r *GormTenantRepository) CreateUsage(ctx context.Context, usage *tenant.TenantUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

// GetUsage retrieves usage by tenant and resource type
func (r *GormTenantRepository) GetUsage(ctx context.Context, tenantID, resourceType string) (*tenant.TenantUsage, error) {
	var usage tenant.TenantUsage
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND resource_type = ?", tenantID, resourceType).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// UpdateUsage updates a usage record
func (r *GormTenantRepository) UpdateUsage(ctx context.Context, usage *tenant.TenantUsage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

// ListUsage lists usage records with filtering
func (r *GormTenantRepository) ListUsage(ctx context.Context, filter *tenant.TenantUsageFilter) ([]*tenant.TenantUsage, error) {
	query := r.db.WithContext(ctx).Model(&tenant.TenantUsage{})

	// Apply filters
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if !filter.PeriodStart.IsZero() {
		query = query.Where("period_start >= ?", filter.PeriodStart)
	}
	if !filter.PeriodEnd.IsZero() {
		query = query.Where("period_end <= ?", filter.PeriodEnd)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var usage []*tenant.TenantUsage
	err := query.Find(&usage).Error
	return usage, err
}


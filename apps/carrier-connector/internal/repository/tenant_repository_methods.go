package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TenantAwareQuery adds tenant filtering to database queries
func (r *GormTenantRepository) TenantAwareQuery(ctx context.Context, model any, tenantID string) *gorm.DB {
	query := r.db.WithContext(ctx).Model(model)

	// Add tenant filter if the model has tenant_id field
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query
}

// EnsureTenantIsolation ensures that queries are tenant-isolated
func (r *GormTenantRepository) EnsureTenantIsolation(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required for tenant-isolated operations")
	}

	// Validate tenant exists
	_, err := r.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	return nil
}

// GetTenantFromContext extracts tenant ID from context
func (r *GormTenantRepository) GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// WithTenantIsolation applies tenant isolation to a database operation
func (r *GormTenantRepository) WithTenantIsolation(ctx context.Context, operation func(*gorm.DB) error) error {
	tenantID := r.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Validate tenant exists
	if err := r.EnsureTenantIsolation(ctx, tenantID); err != nil {
		return err
	}

	// Execute operation with tenant filtering
	tx := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	return operation(tx)
}

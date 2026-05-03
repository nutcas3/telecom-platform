package repository

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mvno"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	// Tenant operations
	CreateTenant(ctx context.Context, tenant *tenant.Tenant) error
	GetTenant(ctx context.Context, id string) (*tenant.Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*tenant.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *tenant.Tenant) error
	DeleteTenant(ctx context.Context, id string) error
	ListTenants(ctx context.Context, filter *tenant.TenantFilter) ([]*tenant.Tenant, error)
	CountTenants(ctx context.Context, filter *tenant.TenantFilter) (int, error)

	// Tenant user operations
	CreateTenantUser(ctx context.Context, user *tenant.TenantUser) error
	GetTenantUser(ctx context.Context, tenantID, userID string) (*tenant.TenantUser, error)
	UpdateTenantUser(ctx context.Context, user *tenant.TenantUser) error
	DeleteTenantUser(ctx context.Context, tenantID, userID string) error
	ListTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) ([]*tenant.TenantUser, error)
	CountTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) (int, error)

	// API key operations
	CreateAPIKey(ctx context.Context, apiKey *tenant.TenantAPIKey) error
	GetAPIKey(ctx context.Context, id string) (*tenant.TenantAPIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*tenant.TenantAPIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *tenant.TenantAPIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
	ListAPIKeys(ctx context.Context, tenantID string) ([]*tenant.TenantAPIKey, error)

	// Usage operations
	CreateUsage(ctx context.Context, usage *tenant.TenantUsage) error
	GetUsage(ctx context.Context, tenantID, resourceType string) (*tenant.TenantUsage, error)
	UpdateUsage(ctx context.Context, usage *tenant.TenantUsage) error
	ListUsage(ctx context.Context, filter *tenant.TenantUsageFilter) ([]*tenant.TenantUsage, error)
	GetUsageStats(ctx context.Context, tenantID string) (*tenant.TenantUsageStats, error)

	// Configuration operations
	GetConfig(ctx context.Context, tenantID string) (*tenant.TenantConfig, error)
	UpdateConfig(ctx context.Context, config *tenant.TenantConfig) error

	// Event operations
	CreateEvent(ctx context.Context, event *tenant.TenantEvent) error
	ListEvents(ctx context.Context, tenantID string, limit int) ([]*tenant.TenantEvent, error)
}

// Repository defines the interface for rate plan data operations
type Repository interface {
	// Rate Plan operations
	CreateRatePlan(ctx context.Context, plan *RatePlan) error
	GetRatePlan(ctx context.Context, id string) (*RatePlan, error)
	UpdateRatePlan(ctx context.Context, plan *RatePlan) error
	DeleteRatePlan(ctx context.Context, id string) error
	ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error)

	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error)
	UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error)
	ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error)

	// Usage operations
	CreateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetUsage(ctx context.Context, id string) (*RatePlanUsage, error)
	UpdateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetCurrentUsage(ctx context.Context, profileID string) (*RatePlanUsage, error)
	ListUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error)

	// Analytics operations
	GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error)
	GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error)
	GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error)

	CreateMVNO(ctx context.Context, mvno *mvno.MVNO) error
	GetMVNO(ctx context.Context, id string) (*mvno.MVNO, error)
	GetMVNOByBusinessID(ctx context.Context, businessID string) (*mvno.MVNO, error)
	UpdateMVNO(ctx context.Context, mvno *mvno.MVNO) error
	ListMVNOs(ctx context.Context, filter *mvno.MVNOFilter) ([]*mvno.MVNO, error)
	DeleteMVNO(ctx context.Context, id string) error
	UpdateMVNOStatus(ctx context.Context, id string, status mvno.MVNOStatus) error
	GetMVNOStats(ctx context.Context) (map[string]any, error)
}

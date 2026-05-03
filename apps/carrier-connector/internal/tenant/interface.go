package tenant

import (
	"context"
	"time"
)

// Repository defines the interface for tenant data operations
type Repository interface {
	// Tenant operations
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, id string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, id string) error
	ListTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error)
	CountTenants(ctx context.Context, filter *TenantFilter) (int, error)

	// Tenant user operations
	CreateTenantUser(ctx context.Context, user *TenantUser) error
	GetTenantUser(ctx context.Context, tenantID, userID string) (*TenantUser, error)
	UpdateTenantUser(ctx context.Context, user *TenantUser) error
	DeleteTenantUser(ctx context.Context, tenantID, userID string) error
	ListTenantUsers(ctx context.Context, filter *TenantUserFilter) ([]*TenantUser, error)
	CountTenantUsers(ctx context.Context, filter *TenantUserFilter) (int, error)

	// API key operations
	CreateAPIKey(ctx context.Context, apiKey *TenantAPIKey) error
	GetAPIKey(ctx context.Context, id string) (*TenantAPIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*TenantAPIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *TenantAPIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
	ListAPIKeys(ctx context.Context, tenantID string) ([]*TenantAPIKey, error)

	// Usage tracking operations
	CreateUsage(ctx context.Context, usage *TenantUsage) error
	GetUsage(ctx context.Context, tenantID, resourceType string) (*TenantUsage, error)
	UpdateUsage(ctx context.Context, usage *TenantUsage) error
	ListUsage(ctx context.Context, filter *TenantUsageFilter) ([]*TenantUsage, error)
	GetUsageStats(ctx context.Context, tenantID string) (*TenantUsageStats, error)

	// Configuration operations
	GetConfig(ctx context.Context, tenantID string) (*TenantConfig, error)
	UpdateConfig(ctx context.Context, config *TenantConfig) error

	// Event operations
	CreateEvent(ctx context.Context, event *TenantEvent) error
	ListEvents(ctx context.Context, tenantID string, limit int) ([]*TenantEvent, error)
}

// Service defines the interface for tenant business operations
type Service interface {
	// Tenant management
	CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error)
	GetTenant(ctx context.Context, id string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, id string, req *UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, id string) error
	ListTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error)

	// User management
	AddUserToTenant(ctx context.Context, req *CreateTenantUserRequest) (*TenantUser, error)
	GetTenantUser(ctx context.Context, tenantID, userID string) (*TenantUser, error)
	UpdateTenantUser(ctx context.Context, tenantID, userID string, req *UpdateTenantUserRequest) (*TenantUser, error)
	RemoveUserFromTenant(ctx context.Context, tenantID, userID string) error
	ListTenantUsers(ctx context.Context, filter *TenantUserFilter) ([]*TenantUser, error)

	// API key management
	CreateAPIKey(ctx context.Context, tenantID string, req *CreateAPIKeyRequest) (*TenantAPIKey, string, error)
	GetAPIKey(ctx context.Context, id string) (*TenantAPIKey, error)
	UpdateAPIKey(ctx context.Context, id string, req *UpdateAPIKeyRequest) (*TenantAPIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error
	ListAPIKeys(ctx context.Context, tenantID string) ([]*TenantAPIKey, error)
	ValidateAPIKey(ctx context.Context, key string) (*TenantAPIKey, error)

	// Quota and usage management
	CheckQuota(ctx context.Context, tenantID, resourceType string, count int) error
	GetUsageStats(ctx context.Context, tenantID string) (*TenantUsageStats, error)
	UpdateUsage(ctx context.Context, tenantID, resourceType string, count int) error
	GetQuotaStatus(ctx context.Context, tenantID string) (map[string]QuotaStatus, error)

	// Configuration management
	GetTenantConfig(ctx context.Context, tenantID string) (*TenantConfig, error)
	UpdateTenantConfig(ctx context.Context, tenantID string, config *TenantConfig) error
	GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error)

	// Authentication and authorization
	ValidateTenantAccess(ctx context.Context, tenantID, userID string) (*TenantContext, error)
	HasPermission(ctx context.Context, tenantID, userID string, permission string) (bool, error)
	GetTenantContext(ctx context.Context, tenantID string) (*TenantContext, error)

	// Analytics and monitoring
	GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error)
	GetTenantEvents(ctx context.Context, tenantID string, limit int) ([]*TenantEvent, error)
	LogTenantEvent(ctx context.Context, event *TenantEvent) error
}

// Middleware defines the interface for tenant middleware
type Middleware interface {
	// Request middleware
	ExtractTenantFromRequest(ctx context.Context, request any) (*TenantContext, error)
	ValidateTenantAccess(ctx context.Context, tenantCtx *TenantContext) error
	InjectTenantContext(ctx context.Context, tenantCtx *TenantContext) context.Context

	// Rate limiting
	CheckRateLimit(ctx context.Context, tenantCtx *TenantContext, endpoint string) error
	RecordAPIUsage(ctx context.Context, tenantCtx *TenantContext, endpoint string)

	// Resource isolation
	IsolateTenantData(ctx context.Context, tenantCtx *TenantContext) context.Context
	ValidateResourceAccess(ctx context.Context, tenantCtx *TenantContext, resource string, resourceID string) error
}

// EventPublisher defines the interface for publishing tenant events
type EventPublisher interface {
	PublishTenantEvent(ctx context.Context, event *TenantEvent) error
	PublishQuotaExceeded(ctx context.Context, tenantID, resourceType string) error
	PublishTenantCreated(ctx context.Context, tenant *Tenant) error
	PublishTenantDeleted(ctx context.Context, tenantID string) error
}

// RateLimiter defines the interface for tenant rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, tenantID, key string) bool
	GetRemaining(ctx context.Context, tenantID, key string) int
	GetResetTime(ctx context.Context, tenantID, key string) time.Time
	Reset(ctx context.Context, tenantID, key string)
}

// ResourceManager defines the interface for managing tenant resources
type ResourceManager interface {
	AllocateResource(ctx context.Context, tenantID, resourceType string, count int) error
	DeallocateResource(ctx context.Context, tenantID, resourceType string, count int) error
	GetResourceUsage(ctx context.Context, tenantID, resourceType string) (*ResourceUsage, error)
	GetResourceQuota(ctx context.Context, tenantID, resourceType string) (*ResourceQuota, error)
	SetResourceQuota(ctx context.Context, tenantID, resourceType string, quota *ResourceQuota) error
}

// ConfigManager defines the interface for managing tenant configuration
type ConfigManager interface {
	GetConfig(ctx context.Context, tenantID string) (*TenantConfig, error)
	SetConfig(ctx context.Context, tenantID string, config *TenantConfig) error
	UpdateConfig(ctx context.Context, tenantID string, updates map[string]any) error
	GetSetting(ctx context.Context, tenantID, key string) (any, error)
	SetSetting(ctx context.Context, tenantID, key string, value any) error
	DeleteSetting(ctx context.Context, tenantID, key string) error
}

// AuditLogger defines the interface for tenant audit logging
type AuditLogger interface {
	LogTenantAction(ctx context.Context, tenantID, userID, action string, details map[string]any) error
	LogAPIAccess(ctx context.Context, tenantID, userID, apiKey, endpoint, method string) error
	LogResourceAccess(ctx context.Context, tenantID, userID, resource, resourceID, action string) error
	LogQuotaViolation(ctx context.Context, tenantID, resourceType string, usage, limit int) error
}

// MetricsCollector defines the interface for collecting tenant metrics
type MetricsCollector interface {
	RecordTenantMetric(ctx context.Context, tenantID, metric string, value float64, tags map[string]string) error
	RecordAPIRequest(ctx context.Context, tenantID, endpoint, method, statusCode string, duration time.Duration) error
	RecordResourceUsage(ctx context.Context, tenantID, resourceType string, count int) error
	RecordUserActivity(ctx context.Context, tenantID, userID, activity string) error
	GetTenantMetrics(ctx context.Context, tenantID string, timeRange string) (map[string]float64, error)
}

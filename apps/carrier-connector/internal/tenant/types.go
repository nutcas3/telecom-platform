package tenant

import (
	"time"
)

// TenantFilter defines filtering options for tenant queries
type TenantFilter struct {
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Domain    string       `json:"domain,omitempty"`
	Status    TenantStatus `json:"status,omitempty"`
	Plan      TenantPlan   `json:"plan,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
	SortBy    string       `json:"sort_by,omitempty"`
	SortOrder string       `json:"sort_order,omitempty"`
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name        string          `json:"name" binding:"required"`
	Domain      string          `json:"domain" binding:"required"`
	Plan        TenantPlan      `json:"plan" binding:"required"`
	MaxUsers    int             `json:"max_users"`
	MaxProfiles int             `json:"max_profiles"`
	MaxCarriers int             `json:"max_carriers"`
	Settings    *TenantSettings `json:"settings,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name        *string         `json:"name,omitempty"`
	Status      *TenantStatus   `json:"status,omitempty"`
	Plan        *TenantPlan     `json:"plan,omitempty"`
	MaxUsers    *int            `json:"max_users,omitempty"`
	MaxProfiles *int            `json:"max_profiles,omitempty"`
	MaxCarriers *int            `json:"max_carriers,omitempty"`
	Settings    *TenantSettings `json:"settings,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// TenantUserFilter defines filtering options for tenant user queries
type TenantUserFilter struct {
	TenantID string           `json:"tenant_id,omitempty"`
	UserID   string           `json:"user_id,omitempty"`
	Email    string           `json:"email,omitempty"`
	Role     TenantRole       `json:"role,omitempty"`
	Status   TenantUserStatus `json:"status,omitempty"`
	Limit    int              `json:"limit,omitempty"`
	Offset   int              `json:"offset,omitempty"`
}

// CreateTenantUserRequest represents a request to add a user to a tenant
type CreateTenantUserRequest struct {
	TenantID string     `json:"tenant_id" binding:"required"`
	UserID   string     `json:"user_id" binding:"required"`
	Email    string     `json:"email" binding:"required,email"`
	Role     TenantRole `json:"role" binding:"required"`
}

// UpdateTenantUserRequest represents a request to update a tenant user
type UpdateTenantUserRequest struct {
	Role   *TenantRole       `json:"role,omitempty"`
	Status *TenantUserStatus `json:"status,omitempty"`
}

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Permissions []string   `json:"permissions"`
	RateLimit   int        `json:"rate_limit"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	Name        *string       `json:"name,omitempty"`
	Permissions []string      `json:"permissions,omitempty"`
	RateLimit   *int          `json:"rate_limit,omitempty"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty"`
	Status      *APIKeyStatus `json:"status,omitempty"`
}

// TenantUsageFilter defines filtering options for tenant usage queries
type TenantUsageFilter struct {
	TenantID     string    `json:"tenant_id,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// TenantUsageStats represents usage statistics for a tenant
type TenantUsageStats struct {
	TenantID          string                 `json:"tenant_id"`
	TotalUsers        int                    `json:"total_users"`
	ActiveUsers       int                    `json:"active_users"`
	TotalProfiles     int                    `json:"total_profiles"`
	ActiveProfiles    int                    `json:"active_profiles"`
	TotalCarriers     int                    `json:"total_carriers"`
	ActiveCarriers    int                    `json:"active_carriers"`
	APIRequests       int64                  `json:"api_requests"`
	StorageUsed       int64                  `json:"storage_used"`
	LastActivity      time.Time              `json:"last_activity"`
	ResourceBreakdown map[string]int64       `json:"resource_breakdown"`
	QuotaStatus       map[string]QuotaStatus `json:"quota_status"`
}

// QuotaStatus represents the status of a resource quota
type QuotaStatus struct {
	Used      int     `json:"used"`
	Limit     int     `json:"limit"`
	Remaining int     `json:"remaining"`
	Percent   float64 `json:"percent"`
	Warning   bool    `json:"warning"`
	Critical  bool    `json:"critical"`
}

// TenantContext represents tenant context for request processing
type TenantContext struct {
	TenantID   string          `json:"tenant_id"`
	TenantName string          `json:"tenant_name"`
	Plan       TenantPlan      `json:"plan"`
	UserID     string          `json:"user_id"`
	UserRole   TenantRole      `json:"user_role"`
	Settings   *TenantSettings `json:"settings"`
	Metadata   map[string]any  `json:"metadata"`
}

// ResourceQuota represents resource quota configuration
type ResourceQuota struct {
	ResourceType     string  `json:"resource_type"`
	Limit            int     `json:"limit"`
	Period           string  `json:"period"` // daily, monthly, yearly
	SoftLimit        bool    `json:"soft_limit"`
	WarningThreshold float64 `json:"warning_threshold"` // percentage
}

// ResourceUsage represents actual resource usage
type ResourceUsage struct {
	ResourceType string    `json:"resource_type"`
	Count        int       `json:"count"`
	LastUpdated  time.Time `json:"last_updated"`
}

// TenantConfig represents tenant-specific configuration
type TenantConfig struct {
	TenantID string          `json:"tenant_id"`
	Config   map[string]any  `json:"config"`
	Settings *TenantSettings `json:"settings"`
	Quotas   []ResourceQuota `json:"quotas"`
	Features map[string]bool `json:"features"`
}

// TenantEvent represents events related to a tenant
type TenantEvent struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	UserID    string          `json:"user_id"`
	EventType TenantEventType `json:"event_type"`
	EventData map[string]any  `json:"event_data"`
	Timestamp time.Time       `json:"timestamp"`
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
}

// TenantEventType represents types of tenant events
type TenantEventType string

const (
	TenantEventCreated       TenantEventType = "tenant_created"
	TenantEventUpdated       TenantEventType = "tenant_updated"
	TenantEventDeleted       TenantEventType = "tenant_deleted"
	TenantEventUserAdded     TenantEventType = "user_added"
	TenantEventUserRemoved   TenantEventType = "user_removed"
	TenantEventUserUpdated   TenantEventType = "user_updated"
	TenantEventAPIKeyCreated TenantEventType = "api_key_created"
	TenantEventAPIKeyRevoked TenantEventType = "api_key_revoked"
	TenantEventQuotaExceeded TenantEventType = "quota_exceeded"
	TenantEventQuotaWarning  TenantEventType = "quota_warning"
	TenantEventLogin         TenantEventType = "login"
	TenantEventLogout        TenantEventType = "logout"
)

// TenantMetrics represents metrics for monitoring tenant health
type TenantMetrics struct {
	TenantID      string    `json:"tenant_id"`
	ActiveUsers   int       `json:"active_users"`
	TotalRequests int64     `json:"total_requests"`
	ErrorRate     float64   `json:"error_rate"`
	ResponseTime  float64   `json:"response_time"`
	StorageUsed   int64     `json:"storage_used"`
	LastActivity  time.Time `json:"last_activity"`
	HealthScore   float64   `json:"health_score"`
	Alerts        []string  `json:"alerts"`
}

// TenantDashboard represents tenant dashboard data
type TenantDashboard struct {
	TenantID     string            `json:"tenant_id"`
	UsageStats   *TenantUsageStats `json:"usage_stats"`
	Metrics      *TenantMetrics    `json:"metrics"`
	RecentEvents []*TenantEvent    `json:"recent_events"`
	QuotaStatus  []*TenantUsage    `json:"quota_status"`
	LastUpdated  time.Time         `json:"last_updated"`
}

// TenantUsageAnalytics represents usage analytics
type TenantUsageAnalytics struct {
	TenantID    string                             `json:"tenant_id"`
	TimeRange   string                             `json:"time_range"`
	StartDate   time.Time                          `json:"start_date"`
	EndDate     time.Time                          `json:"end_date"`
	UsageByType map[string]*ResourceUsageAnalytics `json:"usage_by_type"`
	Trends      map[string][]*UsageTrend           `json:"trends"`
	Peaks       map[string]*UsagePeak              `json:"peaks"`
}

// ResourceUsageAnalytics represents analytics for a specific resource type
type ResourceUsageAnalytics struct {
	ResourceType string    `json:"resource_type"`
	TotalUsage   int       `json:"total_usage"`
	AverageUsage int       `json:"average_usage"`
	PeakUsage    int       `json:"peak_usage"`
	PeakTime     time.Time `json:"peak_time"`
}

// UsageTrend represents usage trend over time
type UsageTrend struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     int       `json:"usage"`
}

// UsagePeak represents a usage peak
type UsagePeak struct {
	Timestamp time.Time      `json:"timestamp"`
	Usage     int            `json:"usage"`
	Context   map[string]any `json:"context"`
}

// TenantPerformanceAnalytics represents performance analytics
type TenantPerformanceAnalytics struct {
	TenantID            string                          `json:"tenant_id"`
	TimeRange           string                          `json:"time_range"`
	StartDate           time.Time                       `json:"start_date"`
	EndDate             time.Time                       `json:"end_date"`
	APIPerformance      *APIPerformance                 `json:"api_performance"`
	ResourcePerformance map[string]*ResourcePerformance `json:"resource_performance"`
	Errors              []*ErrorEvent                   `json:"errors"`
	SlowQueries         []*SlowQuery                    `json:"slow_queries"`
}

// APIPerformance represents API performance metrics
type APIPerformance struct {
	TotalRequests       int     `json:"total_requests"`
	AverageResponseTime float64 `json:"average_response_time"`
	P95ResponseTime     float64 `json:"p95_response_time"`
	ErrorRate           float64 `json:"error_rate"`
	RequestsPerSecond   float64 `json:"requests_per_second"`
}

// ResourcePerformance represents performance metrics for a resource
type ResourcePerformance struct {
	ResourceType string  `json:"resource_type"`
	ResponseTime float64 `json:"response_time"`
	ErrorRate    float64 `json:"error_rate"`
	Throughput   float64 `json:"throughput"`
}

// APIRequestEvent represents an API request event
type APIRequestEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time"`
	UserID       string    `json:"user_id"`
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	Timestamp time.Time      `json:"timestamp"`
	Error     string         `json:"error"`
	Context   map[string]any `json:"context"`
	UserID    string         `json:"user_id"`
}

// SlowQuery represents a slow query event
type SlowQuery struct {
	Timestamp time.Time      `json:"timestamp"`
	Query     string         `json:"query"`
	Duration  time.Duration  `json:"duration"`
	Context   map[string]any `json:"context"`
}

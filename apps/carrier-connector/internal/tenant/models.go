package tenant

import (
	"time"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID          string          `json:"id" gorm:"primaryKey;column:id"`
	Name        string          `json:"name" gorm:"column:name;not null"`
	Domain      string          `json:"domain" gorm:"column:domain;uniqueIndex"`
	Status      TenantStatus    `json:"status" gorm:"column:status;not null"`
	Plan        TenantPlan      `json:"plan" gorm:"column:plan;not null"`
	MaxUsers    int             `json:"max_users" gorm:"column:max_users"`
	MaxProfiles int             `json:"max_profiles" gorm:"column:max_profiles"`
	MaxCarriers int             `json:"max_carriers" gorm:"column:max_carriers"`
	Settings    *TenantSettings `json:"settings" gorm:"column:settings;serializer:json"`
	Metadata    map[string]any  `json:"metadata" gorm:"column:metadata;serializer:json"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

// TableName returns the table name for Tenant
func (Tenant) TableName() string {
	return "tenants"
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantPlan represents the subscription plan for a tenant
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanBasic      TenantPlan = "basic"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// TenantSettings contains tenant-specific configuration
type TenantSettings struct {
	// Currency settings
	DefaultCurrency     string   `json:"default_currency"`
	SupportedCurrencies []string `json:"supported_currencies"`

	// Feature flags
	EnableMultiCurrency     bool `json:"enable_multi_currency"`
	EnableAdvancedAnalytics bool `json:"enable_advanced_analytics"`
	EnableAPIAccess         bool `json:"enable_api_access"`
	EnableWebhooks          bool `json:"enable_webhooks"`

	// Rate limiting
	APIRateLimitPerMinute int `json:"api_rate_limit_per_minute"`
	APIRateLimitPerHour   int `json:"api_rate_limit_per_hour"`

	// Security settings
	Require2FA     bool            `json:"require_2fa"`
	SessionTimeout int             `json:"session_timeout"` // in minutes
	PasswordPolicy *PasswordPolicy `json:"password_policy"`

	// Compliance settings
	DataRetentionDays int      `json:"data_retention_days"`
	ComplianceRegions []string `json:"compliance_regions"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	MaxAgeDays       int  `json:"max_age_days"`
}

// TenantUser represents a user belonging to a tenant
type TenantUser struct {
	ID        string           `json:"id" gorm:"primaryKey;column:id"`
	TenantID  string           `json:"tenant_id" gorm:"column:tenant_id;index"`
	UserID    string           `json:"user_id" gorm:"column:user_id;index"`
	Email     string           `json:"email" gorm:"column:email;not null"`
	Role      TenantRole       `json:"role" gorm:"column:role;not null"`
	Status    TenantUserStatus `json:"status" gorm:"column:status;not null"`
	LastLogin *time.Time       `json:"last_login,omitempty" gorm:"column:last_login"`
	CreatedAt time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for TenantUser
func (TenantUser) TableName() string {
	return "tenant_users"
}

// TenantRole represents the role of a user within a tenant
type TenantRole string

const (
	TenantRoleOwner   TenantRole = "owner"
	TenantRoleAdmin   TenantRole = "admin"
	TenantRoleManager TenantRole = "manager"
	TenantRoleUser    TenantRole = "user"
	TenantRoleViewer  TenantRole = "viewer"
)

// TenantUserStatus represents the status of a tenant user
type TenantUserStatus string

const (
	TenantUserStatusActive    TenantUserStatus = "active"
	TenantUserStatusInactive  TenantUserStatus = "inactive"
	TenantUserStatusSuspended TenantUserStatus = "suspended"
	TenantUserStatusInvited   TenantUserStatus = "invited"
)

// TenantUsage tracks resource usage per tenant
type TenantUsage struct {
	ID             string    `json:"id" gorm:"primaryKey;column:id"`
	TenantID       string    `json:"tenant_id" gorm:"column:tenant_id;index"`
	ResourceType   string    `json:"resource_type" gorm:"column:resource_type"`
	ResourceCount  int       `json:"resource_count" gorm:"column:resource_count"`
	QuotaLimit     int       `json:"quota_limit" gorm:"column:quota_limit"`
	QuotaUsed      int       `json:"quota_used" gorm:"column:quota_used"`
	QuotaRemaining int       `json:"quota_remaining" gorm:"column:quota_remaining"`
	PeriodStart    time.Time `json:"period_start" gorm:"column:period_start"`
	PeriodEnd      time.Time `json:"period_end" gorm:"column:period_end"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for TenantUsage
func (TenantUsage) TableName() string {
	return "tenant_usage"
}

// TenantAPIKey represents API keys for tenant access
type TenantAPIKey struct {
	ID          string       `json:"id" gorm:"primaryKey;column:id"`
	TenantID    string       `json:"tenant_id" gorm:"column:tenant_id;index"`
	Name        string       `json:"name" gorm:"column:name;not null"`
	KeyHash     string       `json:"-" gorm:"column:key_hash;not null"` // hashed API key
	KeyPrefix   string       `json:"key_prefix" gorm:"column:key_prefix;not null"`
	Permissions []string     `json:"permissions" gorm:"column:permissions;serializer:json"`
	RateLimit   int          `json:"rate_limit" gorm:"column:rate_limit"` // requests per minute
	LastUsed    *time.Time   `json:"last_used,omitempty" gorm:"column:last_used"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty" gorm:"column:expires_at"`
	Status      APIKeyStatus `json:"status" gorm:"column:status;not null"`
	CreatedBy   string       `json:"created_by" gorm:"column:created_by"`
	CreatedAt   time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for TenantAPIKey
func (TenantAPIKey) TableName() string {
	return "tenant_api_keys"
}

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusInactive APIKeyStatus = "inactive"
	APIKeyStatusExpired  APIKeyStatus = "expired"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
)

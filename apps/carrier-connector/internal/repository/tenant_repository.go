package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"gorm.io/gorm"
)

// GormTenantRepository implements the tenant repository interface using GORM
type GormTenantRepository struct {
	db *gorm.DB
}

// NewGormTenantRepository creates a new GORM tenant repository
func NewGormTenantRepository(db *gorm.DB) Repository {
	return &GormTenantRepository{db: db}
}

// CreateTenant creates a new tenant
func (r *GormTenantRepository) CreateTenant(ctx context.Context, tenant *Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

// GetTenant retrieves a tenant by ID
func (r *GormTenantRepository) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (r *GormTenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// UpdateTenant updates an existing tenant
func (r *GormTenantRepository) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// DeleteTenant deletes a tenant
func (r *GormTenantRepository) DeleteTenant(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Tenant{}, "id = ?", id).Error
}

// ListTenants lists tenants with filtering
func (r *GormTenantRepository) ListTenants(ctx context.Context, filter *tenant.TenantFilter) ([]*Tenant, error) {
	query := r.db.WithContext(ctx).Model(&Tenant{})

	// Apply filters
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

	var tenants []*Tenant
	err := query.Find(&tenants).Error
	return tenants, err
}

// CountTenants counts tenants with filtering
func (r *GormTenantRepository) CountTenants(ctx context.Context, filter *tenant.TenantFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&Tenant{})

	// Apply filters
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

// CreateTenantUser creates a new tenant user
func (r *GormTenantRepository) CreateTenantUser(ctx context.Context, user *TenantUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetTenantUser retrieves a tenant user
func (r *GormTenantRepository) GetTenantUser(ctx context.Context, tenantID, userID string) (*TenantUser, error) {
	var user TenantUser
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateTenantUser updates a tenant user
func (r *GormTenantRepository) UpdateTenantUser(ctx context.Context, user *TenantUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteTenantUser deletes a tenant user
func (r *GormTenantRepository) DeleteTenantUser(ctx context.Context, tenantID, userID string) error {
	return r.db.WithContext(ctx).Delete(&TenantUser{}, "tenant_id = ? AND user_id = ?", tenantID, userID).Error
}

// ListTenantUsers lists tenant users with filtering
func (r *GormTenantRepository) ListTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) ([]*TenantUser, error) {
	query := r.db.WithContext(ctx).Model(&TenantUser{})

	// Apply filters
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

	var users []*TenantUser
	err := query.Find(&users).Error
	return users, err
}

// CountTenantUsers counts tenant users with filtering
func (r *GormTenantRepository) CountTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&TenantUser{})

	// Apply filters
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

// CreateAPIKey creates a new API key
func (r *GormTenantRepository) CreateAPIKey(ctx context.Context, apiKey *TenantAPIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetAPIKey retrieves an API key by ID
func (r *GormTenantRepository) GetAPIKey(ctx context.Context, id string) (*TenantAPIKey, error) {
	var apiKey TenantAPIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetAPIKeyByHash retrieves an API key by hash
func (r *GormTenantRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*TenantAPIKey, error) {
	var apiKey TenantAPIKey
	err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateAPIKey updates an API key
func (r *GormTenantRepository) UpdateAPIKey(ctx context.Context, apiKey *TenantAPIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// DeleteAPIKey deletes an API key
func (r *GormTenantRepository) DeleteAPIKey(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TenantAPIKey{}, "id = ?", id).Error
}

// ListAPIKeys lists API keys for a tenant
func (r *GormTenantRepository) ListAPIKeys(ctx context.Context, tenantID string) ([]*TenantAPIKey, error) {
	var apiKeys []*TenantAPIKey
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&apiKeys).Error
	return apiKeys, err
}

// CreateUsage creates a new usage record
func (r *GormTenantRepository) CreateUsage(ctx context.Context, usage *TenantUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

// GetUsage retrieves usage by tenant and resource type
func (r *GormTenantRepository) GetUsage(ctx context.Context, tenantID, resourceType string) (*TenantUsage, error) {
	var usage TenantUsage
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND resource_type = ?", tenantID, resourceType).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// UpdateUsage updates a usage record
func (r *GormTenantRepository) UpdateUsage(ctx context.Context, usage *TenantUsage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

// ListUsage lists usage records with filtering
func (r *GormTenantRepository) ListUsage(ctx context.Context, filter *tenant.TenantUsageFilter) ([]*TenantUsage, error) {
	query := r.db.WithContext(ctx).Model(&TenantUsage{})

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

	var usage []*TenantUsage
	err := query.Find(&usage).Error
	return usage, err
}

// GetUsageStats retrieves usage statistics for a tenant
func (r *GormTenantRepository) GetUsageStats(ctx context.Context, tenantID string) (*tenant.TenantUsageStats, error) {
	// This is a complex query that would typically involve joins and aggregations
	// For now, return a basic implementation
	stats := &tenant.TenantUsageStats{
		TenantID:          tenantID,
		ResourceBreakdown: make(map[string]int64),
		QuotaStatus:       make(map[string]tenant.QuotaStatus),
	}

	// Get all usage records for the tenant
	usageRecords, err := r.ListUsage(ctx, &tenant.TenantUsageFilter{TenantID: tenantID})
	if err != nil {
		return nil, err
	}

	// Process usage records
	for _, usage := range usageRecords {
		stats.ResourceBreakdown[usage.ResourceType] = int64(usage.QuotaUsed)
		stats.QuotaStatus[usage.ResourceType] = tenant.QuotaStatus{
			Used:      usage.QuotaUsed,
			Limit:     usage.QuotaLimit,
			Remaining: usage.QuotaRemaining,
		}
	}

	return stats, nil
}

// GetConfig retrieves tenant configuration
func (r *GormTenantRepository) GetConfig(ctx context.Context, tenantID string) (*TenantConfig, error) {
	// Get tenant to extract settings
	tenant, err := r.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Create basic config
	config := &TenantConfig{
		TenantID: tenantID,
		Config:   make(map[string]interface{}),
		Settings: tenant.Settings,
		Quotas:   []ResourceQuota{},
		Features: make(map[string]bool),
	}

	// Add default quotas based on plan
	switch tenant.Plan {
	case TenantPlanFree:
		config.Quotas = []ResourceQuota{
			{ResourceType: "users", Limit: 5, Period: "monthly"},
			{ResourceType: "profiles", Limit: 100, Period: "monthly"},
			{ResourceType: "carriers", Limit: 3, Period: "monthly"},
		}
	case TenantPlanBasic:
		config.Quotas = []ResourceQuota{
			{ResourceType: "users", Limit: 25, Period: "monthly"},
			{ResourceType: "profiles", Limit: 1000, Period: "monthly"},
			{ResourceType: "carriers", Limit: 10, Period: "monthly"},
		}
	case TenantPlanPro:
		config.Quotas = []ResourceQuota{
			{ResourceType: "users", Limit: 100, Period: "monthly"},
			{ResourceType: "profiles", Limit: 10000, Period: "monthly"},
			{ResourceType: "carriers", Limit: 50, Period: "monthly"},
		}
	case TenantPlanEnterprise:
		config.Quotas = []ResourceQuota{
			{ResourceType: "users", Limit: -1, Period: "monthly"},
			{ResourceType: "profiles", Limit: -1, Period: "monthly"},
			{ResourceType: "carriers", Limit: -1, Period: "monthly"},
		}
	}

	// Add default features based on plan
	switch tenant.Plan {
	case TenantPlanFree:
		config.Features = map[string]bool{
			"multi_currency":     false,
			"advanced_analytics": false,
			"api_access":         true,
			"webhooks":           false,
		}
	case TenantPlanBasic:
		config.Features = map[string]bool{
			"multi_currency":     true,
			"advanced_analytics": false,
			"api_access":         true,
			"webhooks":           false,
		}
	case TenantPlanPro, TenantPlanEnterprise:
		config.Features = map[string]bool{
			"multi_currency":     true,
			"advanced_analytics": true,
			"api_access":         true,
			"webhooks":           true,
		}
	}

	return config, nil
}

// UpdateConfig updates tenant configuration
func (r *GormTenantRepository) UpdateConfig(ctx context.Context, config *TenantConfig) error {
	// Store configuration in tenant metadata or separate table
	// For now, update tenant settings
	tenant, err := r.GetTenant(ctx, config.TenantID)
	if err != nil {
		return err
	}

	tenant.Settings = config.Settings
	tenant.Metadata = config.Config
	tenant.UpdatedAt = time.Now()

	return r.UpdateTenant(ctx, tenant)
}

// CreateEvent creates a new tenant event
func (r *GormTenantRepository) CreateEvent(ctx context.Context, event *TenantEvent) error {
	// Store events in a separate table or log system
	// For now, we'll use a simple approach with JSON storage
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Create a simple event record (in a real implementation, this would be a proper table)
	eventRecord := struct {
		ID        string    `gorm:"primaryKey"`
		TenantID  string    `gorm:"index"`
		EventData string    `gorm:"type:text"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
	}{
		ID:        event.ID,
		TenantID:  event.TenantID,
		EventData: string(eventData),
		CreatedAt: event.Timestamp,
	}

	return r.db.WithContext(ctx).Table("tenant_events").Create(&eventRecord).Error
}

// ListEvents lists tenant events
func (r *GormTenantRepository) ListEvents(ctx context.Context, tenantID string, limit int) ([]*TenantEvent, error) {
	// Query events from the events table
	var eventRecords []struct {
		ID        string    `gorm:"primaryKey"`
		TenantID  string    `gorm:"index"`
		EventData string    `gorm:"type:text"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
	}

	query := r.db.WithContext(ctx).Table("tenant_events").Where("tenant_id = ?", tenantID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("created_at DESC")

	err := query.Find(&eventRecords).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal events
	var events []*TenantEvent
	for _, record := range eventRecords {
		var event TenantEvent
		if err := json.Unmarshal([]byte(record.EventData), &event); err != nil {
			continue // Skip malformed events
		}
		events = append(events, &event)
	}

	return events, nil
}

// Helper methods for tenant isolation in other repositories

// TenantAwareQuery adds tenant filtering to database queries
func (r *GormTenantRepository) TenantAwareQuery(ctx context.Context, model interface{}, tenantID string) *gorm.DB {
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

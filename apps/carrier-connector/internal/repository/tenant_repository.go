package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

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
func (r *GormTenantRepository) GetConfig(ctx context.Context, tenantID string) (*tenant.TenantConfig, error) {
	// Get tenant to extract settings
	tenantRecord, err := r.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Create basic config
	config := &tenant.TenantConfig{
		TenantID: tenantID,
		Config:   make(map[string]any),
		Settings: tenantRecord.Settings,
		Quotas:   []tenant.ResourceQuota{},
		Features: make(map[string]bool),
	}

	// Add default quotas based on plan
	switch tenantRecord.Plan {
	case tenant.TenantPlanFree:
		config.Quotas = []tenant.ResourceQuota{
			{ResourceType: "users", Limit: 5, Period: "monthly"},
			{ResourceType: "profiles", Limit: 100, Period: "monthly"},
			{ResourceType: "carriers", Limit: 3, Period: "monthly"},
		}
	case tenant.TenantPlanBasic:
		config.Quotas = []tenant.ResourceQuota{
			{ResourceType: "users", Limit: 25, Period: "monthly"},
			{ResourceType: "profiles", Limit: 1000, Period: "monthly"},
			{ResourceType: "carriers", Limit: 10, Period: "monthly"},
		}
	case tenant.TenantPlanPro:
		config.Quotas = []tenant.ResourceQuota{
			{ResourceType: "users", Limit: 100, Period: "monthly"},
			{ResourceType: "profiles", Limit: 10000, Period: "monthly"},
			{ResourceType: "carriers", Limit: 50, Period: "monthly"},
		}
	case tenant.TenantPlanEnterprise:
		config.Quotas = []tenant.ResourceQuota{
			{ResourceType: "users", Limit: -1, Period: "monthly"},
			{ResourceType: "profiles", Limit: -1, Period: "monthly"},
			{ResourceType: "carriers", Limit: -1, Period: "monthly"},
		}
	}

	// Add default features based on plan
	switch tenantRecord.Plan {
	case tenant.TenantPlanFree:
		config.Features = map[string]bool{
			"multi_currency":     false,
			"advanced_analytics": false,
			"api_access":         true,
			"webhooks":           false,
		}
	case tenant.TenantPlanBasic:
		config.Features = map[string]bool{
			"multi_currency":     true,
			"advanced_analytics": false,
			"api_access":         true,
			"webhooks":           false,
		}
	case tenant.TenantPlanPro, tenant.TenantPlanEnterprise:
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
func (r *GormTenantRepository) UpdateConfig(ctx context.Context, config *tenant.TenantConfig) error {
	// Store configuration in tenant metadata or separate table
	// For now, update tenant settings
	tenantRecord, err := r.GetTenant(ctx, config.TenantID)
	if err != nil {
		return err
	}

	tenantRecord.Settings = config.Settings
	tenantRecord.Metadata = config.Config
	tenantRecord.UpdatedAt = time.Now()

	return r.UpdateTenant(ctx, tenantRecord)
}

// CreateEvent creates a new tenant event
func (r *GormTenantRepository) CreateEvent(ctx context.Context, event *tenant.TenantEvent) error {
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
func (r *GormTenantRepository) ListEvents(ctx context.Context, tenantID string, limit int) ([]*tenant.TenantEvent, error) {
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
	var events []*tenant.TenantEvent
	for _, record := range eventRecords {
		var event tenant.TenantEvent
		if err := json.Unmarshal([]byte(record.EventData), &event); err != nil {
			continue // Skip malformed events
		}
		events = append(events, &event)
	}

	return events, nil
}

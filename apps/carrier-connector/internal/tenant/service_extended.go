package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// CheckQuota checks if tenant has sufficient quota for a resource
func (s *ServiceImpl) CheckQuota(ctx context.Context, tenantID, resourceType string, count int) error {
	// Get current usage
	usage, err := s.repository.GetUsage(ctx, tenantID, resourceType)
	if err != nil {
		// If no usage record exists, create one
		usage = &TenantUsage{
			ID:             generateID(),
			TenantID:       tenantID,
			ResourceType:   resourceType,
			ResourceCount:  0,
			QuotaUsed:      0,
			QuotaRemaining: 0,
			PeriodStart:    time.Now().Truncate(24 * time.Hour),
			PeriodEnd:      time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	}

	// Get tenant configuration
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Find quota for resource type
	var quotaLimit int
	for _, quota := range config.Quotas {
		if quota.ResourceType == resourceType {
			quotaLimit = quota.Limit
			break
		}
	}

	// Check if quota is unlimited (-1)
	if quotaLimit == -1 {
		return nil
	}

	// Check if adding count would exceed quota
	if usage.QuotaUsed+count > quotaLimit {
		// Log quota exceeded event
		event := &TenantEvent{
			ID:        generateID(),
			TenantID:  tenantID,
			UserID:    "",
			EventType: TenantEventQuotaExceeded,
			EventData: map[string]interface{}{
				"resource_type": resourceType,
				"requested":     count,
				"current_usage": usage.QuotaUsed,
				"quota_limit":   quotaLimit,
			},
			Timestamp: time.Now(),
		}

		if err := s.repository.CreateEvent(ctx, event); err != nil {
			s.logger.WithError(err).Error("Failed to create quota exceeded event")
		}

		return fmt.Errorf("quota exceeded for %s: %d/%d used", resourceType, usage.QuotaUsed, quotaLimit)
	}

	return nil
}

// GetUsageStats retrieves usage statistics for a tenant
func (s *ServiceImpl) GetUsageStats(ctx context.Context, tenantID string) (*TenantUsageStats, error) {
	// Get usage records
	usageFilter := &TenantUsageFilter{
		TenantID: tenantID,
	}

	usageRecords, err := s.repository.ListUsage(ctx, usageFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	// Get tenant configuration
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Get tenant users
	userFilter := &TenantUserFilter{
		TenantID: tenantID,
		Status:   TenantUserStatusActive,
	}

	users, err := s.repository.ListTenantUsers(ctx, userFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}

	// Build usage stats
	stats := &TenantUsageStats{
		TenantID:          tenantID,
		TotalUsers:        len(users),
		ActiveUsers:       len(users),
		ResourceBreakdown: make(map[string]int64),
		QuotaStatus:       make(map[string]QuotaStatus),
	}

	// Process usage records
	for _, usage := range usageRecords {
		stats.ResourceBreakdown[usage.ResourceType] = int64(usage.QuotaUsed)

		// Calculate quota status
		var quotaLimit int
		for _, quota := range config.Quotas {
			if quota.ResourceType == usage.ResourceType {
				quotaLimit = quota.Limit
				break
			}
		}

		quotaStatus := QuotaStatus{
			Used:      usage.QuotaUsed,
			Limit:     quotaLimit,
			Remaining: quotaLimit - usage.QuotaUsed,
		}

		if quotaLimit > 0 {
			quotaStatus.Percent = float64(usage.QuotaUsed) / float64(quotaLimit) * 100
			quotaStatus.Warning = quotaStatus.Percent >= 80
			quotaStatus.Critical = quotaStatus.Percent >= 95
		}

		stats.QuotaStatus[usage.ResourceType] = quotaStatus
	}

	return stats, nil
}

// UpdateUsage updates resource usage for a tenant
func (s *ServiceImpl) UpdateUsage(ctx context.Context, tenantID, resourceType string, count int) error {
	// Get current usage
	usage, err := s.repository.GetUsage(ctx, tenantID, resourceType)
	if err != nil {
		// Create new usage record
		usage = &TenantUsage{
			ID:            generateID(),
			TenantID:      tenantID,
			ResourceType:  resourceType,
			ResourceCount: count,
			QuotaUsed:     count,
			PeriodStart:   time.Now().Truncate(24 * time.Hour),
			PeriodEnd:     time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
	} else {
		// Update existing usage
		usage.ResourceCount += count
		usage.QuotaUsed += count
		usage.UpdatedAt = time.Now()
	}

	// Save usage
	if err := s.repository.UpdateUsage(ctx, usage); err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return nil
}

// GetQuotaStatus retrieves quota status for all tenant resources
func (s *ServiceImpl) GetQuotaStatus(ctx context.Context, tenantID string) (map[string]QuotaStatus, error) {
	// Get usage stats
	stats, err := s.GetUsageStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return stats.QuotaStatus, nil
}

// GetTenantConfig retrieves tenant configuration
func (s *ServiceImpl) GetTenantConfig(ctx context.Context, tenantID string) (*TenantConfig, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config, nil
}

// UpdateTenantConfig updates tenant configuration
func (s *ServiceImpl) UpdateTenantConfig(ctx context.Context, tenantID string, config *TenantConfig) error {
	config.TenantID = tenantID

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update tenant config: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant config updated successfully")

	return nil
}

// GetTenantSettings retrieves tenant settings
func (s *ServiceImpl) GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config.Settings, nil
}

// ValidateTenantAccess validates tenant access
func (s *ServiceImpl) ValidateTenantAccess(ctx context.Context, tenantID, userID string) (*TenantContext, error) {
	// Get tenant
	tenant, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check tenant status
	if tenant.Status != TenantStatusActive {
		return nil, errors.New("tenant is not active")
	}

	// Get tenant configuration
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Get user role if userID provided
	var userRole TenantRole
	if userID != "" {
		user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
		if err != nil {
			return nil, fmt.Errorf("user not found in tenant: %w", err)
		}

		if user.Status != TenantUserStatusActive {
			return nil, errors.New("user is not active in tenant")
		}

		userRole = user.Role
	}

	// Create tenant context
	tenantCtx := &TenantContext{
		TenantID:   tenantID,
		TenantName: tenant.Name,
		Plan:       tenant.Plan,
		UserID:     userID,
		UserRole:   userRole,
		Settings:   config.Settings,
		Metadata:   tenant.Metadata,
	}

	return tenantCtx, nil
}

// HasPermission checks if user has specific permission
func (s *ServiceImpl) HasPermission(ctx context.Context, tenantID, userID string, permission string) (bool, error) {
	// Get tenant user
	user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
	if err != nil {
		return false, fmt.Errorf("user not found in tenant: %w", err)
	}

	// Define permission matrix
	permissionMatrix := map[TenantRole]map[string]bool{
		TenantRoleOwner: {
			"tenant:read":   true,
			"tenant:write":  true,
			"tenant:delete": true,
			"user:read":     true,
			"user:write":    true,
			"user:delete":   true,
			"apikey:read":   true,
			"apikey:write":  true,
			"apikey:delete": true,
			"config:read":   true,
			"config:write":  true,
		},
		TenantRoleAdmin: {
			"tenant:read":   true,
			"tenant:write":  true,
			"user:read":     true,
			"user:write":    true,
			"user:delete":   true,
			"apikey:read":   true,
			"apikey:write":  true,
			"apikey:delete": true,
			"config:read":   true,
			"config:write":  true,
		},
		TenantRoleManager: {
			"tenant:read":  true,
			"user:read":    true,
			"user:write":   true,
			"apikey:read":  true,
			"apikey:write": true,
			"config:read":  true,
		},
		TenantRoleUser: {
			"tenant:read": true,
			"apikey:read": true,
			"config:read": true,
		},
		TenantRoleViewer: {
			"tenant:read": true,
			"apikey:read": true,
			"config:read": true,
		},
	}

	// Check permission
	rolePermissions, exists := permissionMatrix[user.Role]
	if !exists {
		return false, nil
	}

	hasPermission, exists := rolePermissions[permission]
	if !exists {
		return false, nil
	}

	return hasPermission, nil
}

// GetTenantContext retrieves tenant context
func (s *ServiceImpl) GetTenantContext(ctx context.Context, tenantID string) (*TenantContext, error) {
	return s.ValidateTenantAccess(ctx, tenantID, "")
}

// CheckRateLimit checks rate limit for tenant
func (s *ServiceImpl) CheckRateLimit(ctx context.Context, tenantCtx *TenantContext, endpoint string) error {
	// Use rate limiter to check if request is allowed
	key := fmt.Sprintf("%s:%s", tenantCtx.TenantID, endpoint)

	if !s.rateLimiter.Allow(ctx, tenantCtx.TenantID, key) {
		return errors.New("rate limit exceeded")
	}

	return nil
}

// RecordAPIUsage records API usage for rate limiting
func (s *ServiceImpl) RecordAPIUsage(ctx context.Context, tenantCtx *TenantContext, endpoint string) {
	// Record metrics
	if s.metricsCollector != nil {
		s.metricsCollector.RecordAPIRequest(ctx, tenantCtx.TenantID, endpoint, "", "", 0)
	}
}

// ValidateResourceAccess validates resource access
func (s *ServiceImpl) ValidateResourceAccess(ctx context.Context, tenantCtx *TenantContext, resource string, resourceID string) error {
	// For now, implement basic validation
	// In a real implementation, this would check resource ownership
	// and access patterns based on resource type

	switch resource {
	case "tenant":
		if resourceID != tenantCtx.TenantID {
			return errors.New("access denied: tenant mismatch")
		}
	case "user", "apikey", "config":
		// These resources belong to the tenant, so basic tenant validation is sufficient
		return nil
	default:
		// Unknown resource type, deny access
		return errors.New("access denied: unknown resource type")
	}

	return nil
}

// GetTenantMetrics retrieves tenant metrics
func (s *ServiceImpl) GetTenantMetrics(ctx context.Context, tenantID string) (*TenantMetrics, error) {
	// Get usage stats
	usageStats, err := s.GetUsageStats(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// Get recent events
	events, err := s.repository.ListEvents(ctx, tenantID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant events: %w", err)
	}

	// Calculate metrics
	metrics := &TenantMetrics{
		TenantID:    tenantID,
		ActiveUsers: usageStats.ActiveUsers,
		StorageUsed: 0, // Would be calculated from actual storage usage
		HealthScore: 100.0,
		Alerts:      []string{},
	}

	// Calculate last activity
	if len(events) > 0 {
		metrics.LastActivity = events[0].Timestamp
	}

	// Calculate error rate and response time from events
	errorCount := 0
	totalRequests := 0
	var totalResponseTime time.Duration

	for _, event := range events {
		if event.EventType == "api_request" {
			totalRequests++
			if statusCode, exists := event.EventData["status_code"]; exists {
				if code, ok := statusCode.(float64); ok && code >= 400 {
					errorCount++
				}
			}
			if responseTime, exists := event.EventData["response_time"]; exists {
				if rt, ok := responseTime.(float64); ok {
					totalResponseTime += time.Duration(rt) * time.Millisecond
				}
			}
		}
	}

	if totalRequests > 0 {
		metrics.ErrorRate = float64(errorCount) / float64(totalRequests) * 100
		metrics.ResponseTime = float64(totalResponseTime) / float64(totalRequests) / float64(time.Millisecond)
	}

	// Check for alerts
	for resourceType, quotaStatus := range usageStats.QuotaStatus {
		if quotaStatus.Critical {
			metrics.Alerts = append(metrics.Alerts, fmt.Sprintf("Critical: %s quota at %.1f%%", resourceType, quotaStatus.Percent))
			metrics.HealthScore -= 20
		} else if quotaStatus.Warning {
			metrics.Alerts = append(metrics.Alerts, fmt.Sprintf("Warning: %s quota at %.1f%%", resourceType, quotaStatus.Percent))
			metrics.HealthScore -= 10
		}
	}

	return metrics, nil
}

// GetTenantEvents retrieves tenant events
func (s *ServiceImpl) GetTenantEvents(ctx context.Context, tenantID string, limit int) ([]*TenantEvent, error) {
	events, err := s.repository.ListEvents(ctx, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant events: %w", err)
	}

	return events, nil
}

// LogTenantEvent logs a tenant event
func (s *ServiceImpl) LogTenantEvent(ctx context.Context, event *TenantEvent) error {
	if err := s.repository.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create tenant event: %w", err)
	}

	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("tnt_%d", time.Now().UnixNano())
}

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// CheckQuota checks if tenant has sufficient quota for a resource
func (s *TenantServiceImpl) CheckQuota(ctx context.Context, tenantID, resourceType string, amount int) error {
	// Get current usage
	usage, err := s.repository.GetUsage(ctx, tenantID, resourceType)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Get quota limits
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

	// Check if quota is exceeded
	if quotaLimit >= 0 && usage.QuotaUsed+amount > quotaLimit {
		return fmt.Errorf("quota exceeded for resource %s: %d/%d", resourceType, usage.QuotaUsed+amount, quotaLimit)
	}

	return nil
}

// GetUsageStats retrieves usage statistics for a tenant
func (s *TenantServiceImpl) GetUsageStats(ctx context.Context, tenantID string) (*tenant.TenantUsageStats, error) {
	// Get all usage records
	usageRecords, err := s.repository.ListUsage(ctx, &tenant.TenantUsageFilter{TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("failed to list usage: %w", err)
	}

	// Calculate stats
	stats := &tenant.TenantUsageStats{
		TenantID:          tenantID,
		TotalUsers:        0,
		ActiveUsers:       0,
		TotalProfiles:     0,
		ActiveProfiles:    0,
		TotalCarriers:     0,
		ActiveCarriers:    0,
		APIRequests:       0,
		StorageUsed:       0,
		LastActivity:      time.Time{},
		ResourceBreakdown: make(map[string]int64),
		QuotaStatus:       make(map[string]tenant.QuotaStatus),
	}

	// Get tenant config for quotas
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err == nil {
		for _, quota := range config.Quotas {
			// Find current usage for this resource
			var currentUsage int
			for _, usage := range usageRecords {
				if usage.ResourceType == quota.ResourceType {
					currentUsage = usage.QuotaUsed
					break
				}
			}

			// Calculate quota status
			percent := float64(currentUsage) / float64(quota.Limit) * 100
			status := tenant.QuotaStatus{
				Used:      currentUsage,
				Limit:     quota.Limit,
				Remaining: quota.Limit - currentUsage,
				Percent:   percent,
				Warning:   percent >= 80,
				Critical:  percent >= 95,
			}
			stats.QuotaStatus[quota.ResourceType] = status
		}
	}

	return stats, nil
}

// UpdateUsage updates resource usage for a tenant
func (s *TenantServiceImpl) UpdateUsage(ctx context.Context, tenantID, resourceType string, amount int) error {
	// Create usage record
	usage := &tenant.TenantUsage{
		ID:           fmt.Sprintf("usage_%d", time.Now().UnixNano()),
		TenantID:     tenantID,
		ResourceType: resourceType,
		QuotaUsed:    amount,
		UpdatedAt:    time.Now(),
	}

	// Update usage
	if err := s.repository.UpdateUsage(ctx, usage); err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return nil
}

// GetQuotaStatus retrieves quota status for all resources
func (s *TenantServiceImpl) GetQuotaStatus(ctx context.Context, tenantID string) (map[string]tenant.QuotaStatus, error) {
	// Get usage stats
	stats, err := s.GetUsageStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return stats.QuotaStatus, nil
}

// CheckRateLimit checks rate limit for tenant
func (s *TenantServiceImpl) CheckRateLimit(ctx context.Context, tenantCtx *tenant.TenantContext, endpoint string) error {
	// Use rate limiter to check if request is allowed
	key := fmt.Sprintf("%s:%s", tenantCtx.TenantID, endpoint)

	if !s.rateLimiter.Allow(ctx, tenantCtx.TenantID, key) {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// RecordAPIUsage records API usage for a tenant
func (s *TenantServiceImpl) RecordAPIUsage(ctx context.Context, tenantID, endpoint string, statusCode int, responseTime time.Duration) error {
	// Create usage record
	usage := &tenant.TenantUsage{
		ID:           fmt.Sprintf("api_%d", time.Now().UnixNano()),
		TenantID:     tenantID,
		ResourceType: "api_calls",
		QuotaUsed:    1,
		UpdatedAt:    time.Now(),
	}

	// Update usage
	if err := s.repository.UpdateUsage(ctx, usage); err != nil {
		return fmt.Errorf("failed to record API usage: %w", err)
	}

	return nil
}

// ValidateResourceAccess validates access to a specific resource
func (s *TenantServiceImpl) ValidateResourceAccess(ctx context.Context, tenantID, resourceType, resourceID string) error {
	// Get tenant
	tenant, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Check tenant status
	if tenant.Status != "active" {
		return fmt.Errorf("tenant is not active")
	}

	// Check resource access based on type
	switch resourceType {
	case "profiles", "carriers", "users":
		// These resources belong to the tenant, so basic tenant validation is sufficient
		return nil
	default:
		// Unknown resource type, deny access
		return fmt.Errorf("access denied: unknown resource type")
	}
}

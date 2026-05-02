package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

func (s *TenantServiceImpl) GetTenantMetrics(ctx context.Context, tenantID string) (*tenant.TenantMetrics, error) {
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
	metrics := &tenant.TenantMetrics{
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

func (s *TenantServiceImpl) GetTenantEvents(ctx context.Context, tenantID string, limit int) ([]*tenant.TenantEvent, error) {
	events, err := s.repository.ListEvents(ctx, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant events: %w", err)
	}

	return events, nil
}

func (s *TenantServiceImpl) LogTenantEvent(ctx context.Context, event *tenant.TenantEvent) error {
	if err := s.repository.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create tenant event: %w", err)
	}

	return nil
}

func (s *TenantServiceImpl) GetTenantDashboard(ctx context.Context, tenantID string) (*tenant.TenantDashboard, error) {

	usageStats, err := s.GetUsageStats(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	metrics, err := s.GetTenantMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant metrics: %w", err)
	}

	recentEvents, err := s.repository.ListEvents(ctx, tenantID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	quotaStatus := s.buildQuotaStatus(usageStats)

	dashboard := &tenant.TenantDashboard{
		TenantID:     tenantID,
		UsageStats:   usageStats,
		Metrics:      metrics,
		RecentEvents: recentEvents,
		QuotaStatus:  quotaStatus,
		LastUpdated:  time.Now(),
	}

	return dashboard, nil
}

func (s *TenantServiceImpl) GetPerformanceAnalytics(ctx context.Context, tenantID string, timeRange string) (*tenant.TenantPerformanceAnalytics, error) {
	startDate, endDate := s.parseTimeRange(timeRange)

	events, err := s.repository.ListEvents(ctx, tenantID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant events: %w", err)
	}

	apiRequests := s.parseAPIRequestEvents(events)

	analytics := &tenant.TenantPerformanceAnalytics{
		TenantID:            tenantID,
		TimeRange:           timeRange,
		StartDate:           startDate,
		EndDate:             endDate,
		APIPerformance:      s.calculateAPIPerformance(apiRequests),
		ResourcePerformance: s.buildResourcePerformance(ctx, tenantID, timeRange),
		Errors:              s.parseErrorEvents(events),
		SlowQueries:         s.parseSlowQueryEvents(events),
	}

	return analytics, nil
}

func generateUsageID(resourceType, tenantID string) string {
	// TODO: Use resourceType and tenantID in ID generation for better traceability
	// For now, use generic usage prefix - could be enhanced to include resource type
	_ = resourceType // Suppress unused parameter warning until implementation is complete
	_ = tenantID     // Suppress unused parameter warning until implementation is complete
	return id.GeneratePrefixed("usage")
}

func (s *TenantServiceImpl) parseAPIRequestEvents(events []*tenant.TenantEvent) []*tenant.APIRequestEvent {
	apiRequests := make([]*tenant.APIRequestEvent, 0)

	for _, event := range events {
		if event.EventType == "api_request" {
			request := s.parseAPIRequestEvent(event)
			apiRequests = append(apiRequests, request)
		}
	}

	return apiRequests
}

func (s *TenantServiceImpl) buildResourcePerformance(ctx context.Context, tenantID, timeRange string) map[string]*tenant.ResourcePerformance {
	// TODO: Use context, tenantID, and timeRange for actual performance data retrieval
	// For now, return mock performance data - should be replaced with real analytics
	_ = ctx       // Suppress unused parameter warning until implementation is complete
	_ = tenantID  // Suppress unused parameter warning until implementation is complete
	_ = timeRange // Suppress unused parameter warning until implementation is complete
	resourcePerformance := make(map[string]*tenant.ResourcePerformance)

	resourceTypes := []string{"users", "profiles", "carriers", "api_calls", "storage"}

	for _, resourceType := range resourceTypes {
		performance := &tenant.ResourcePerformance{
			ResourceType: resourceType,
			ResponseTime: 150.5,  // Mock response time in ms
			Throughput:   1000.0, // Mock requests per second
			ErrorRate:    2.1,    // Mock error rate percentage
		}

		resourcePerformance[resourceType] = performance
	}

	return resourcePerformance
}

func (s *TenantServiceImpl) parseErrorEvents(events []*tenant.TenantEvent) []*tenant.ErrorEvent {
	errorEvents := make([]*tenant.ErrorEvent, 0)

	for _, event := range events {
		if event.EventType == "error" {
			errorEvent := s.parseErrorEvent(event)
			errorEvents = append(errorEvents, errorEvent)
		}
	}

	return errorEvents
}

func (s *TenantServiceImpl) parseSlowQueryEvents(events []*tenant.TenantEvent) []*tenant.SlowQuery {
	slowQueries := make([]*tenant.SlowQuery, 0)

	for _, event := range events {
		if event.EventType == "slow_query" {
			slowQuery := s.parseSlowQueryEvent(event)
			slowQueries = append(slowQueries, slowQuery)
		}
	}

	return slowQueries
}

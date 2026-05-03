package services

import (
	"context"
	"fmt"
	"time"

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
	resourcePerformance := make(map[string]*tenant.ResourcePerformance)

	// Determine event limit based on time range
	eventLimit := 100
	switch timeRange {
	case "1h":
		eventLimit = 50
	case "24h":
		eventLimit = 200
	case "7d":
		eventLimit = 500
	case "30d":
		eventLimit = 1000
	}

	// Retrieve actual tenant events for performance calculation
	events, err := s.repository.ListEvents(ctx, tenantID, eventLimit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve tenant events for performance analytics")
		return resourcePerformance
	}

	// Aggregate performance metrics per resource type from events
	type perfAccumulator struct {
		totalResponseTime float64
		totalRequests     int
		errorCount        int
	}
	accumulators := make(map[string]*perfAccumulator)

	for _, event := range events {
		resourceType, _ := event.EventData["resource_type"].(string)
		if resourceType == "" {
			resourceType = string(event.EventType)
		}

		acc, exists := accumulators[resourceType]
		if !exists {
			acc = &perfAccumulator{}
			accumulators[resourceType] = acc
		}

		acc.totalRequests++

		if respTime, ok := event.EventData["response_time"].(float64); ok {
			acc.totalResponseTime += respTime
		}

		if event.EventType == "error" || event.EventType == tenant.TenantEventQuotaExceeded {
			acc.errorCount++
		}
	}

	// Convert accumulators to ResourcePerformance
	for resourceType, acc := range accumulators {
		avgResponseTime := 0.0
		if acc.totalRequests > 0 {
			avgResponseTime = acc.totalResponseTime / float64(acc.totalRequests)
		}

		errorRate := 0.0
		if acc.totalRequests > 0 {
			errorRate = (float64(acc.errorCount) / float64(acc.totalRequests)) * 100
		}

		resourcePerformance[resourceType] = &tenant.ResourcePerformance{
			ResourceType: resourceType,
			ResponseTime: avgResponseTime,
			Throughput:   float64(acc.totalRequests),
			ErrorRate:    errorRate,
		}
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

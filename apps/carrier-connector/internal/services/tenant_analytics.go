package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// GetTenantMetrics retrieves tenant metrics
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
		TenantID:     tenantID,
		ActiveUsers:  usageStats.ActiveUsers,
		StorageUsed:  0, // Would be calculated from actual storage usage
		HealthScore:  100.0,
		Alerts:       []string{},
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
func (s *TenantServiceImpl) GetTenantEvents(ctx context.Context, tenantID string, limit int) ([]*tenant.TenantEvent, error) {
	events, err := s.repository.ListEvents(ctx, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant events: %w", err)
	}

	return events, nil
}

// LogTenantEvent logs a tenant event
func (s *TenantServiceImpl) LogTenantEvent(ctx context.Context, event *tenant.TenantEvent) error {
	if err := s.repository.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create tenant event: %w", err)
	}

	return nil
}

// GetTenantDashboard returns dashboard data for a tenant
func (s *TenantServiceImpl) GetTenantDashboard(ctx context.Context, tenantID string) (*tenant.TenantDashboard, error) {
	// TODO: Implement proper dashboard when type conversion issues are resolved
	// For now, return a basic dashboard
	dashboard := &tenant.TenantDashboard{
		TenantID:     tenantID,
		UsageStats:   nil,
		Metrics:      nil,
		RecentEvents: nil,
		QuotaStatus:  nil,
		LastUpdated:  time.Now(),
	}

	return dashboard, nil
}

// GetUsageAnalytics returns detailed usage analytics for a tenant
func (s *TenantServiceImpl) GetUsageAnalytics(ctx context.Context, tenantID string, timeRange string) (*tenant.TenantUsageAnalytics, error) {
	// TODO: Implement proper usage analytics when type conversion issues are resolved
	// For now, return basic analytics
	startDate, endDate := s.parseTimeRange(timeRange)
	
	analytics := &tenant.TenantUsageAnalytics{
		TenantID:    tenantID,
		TimeRange:   timeRange,
		StartDate:   startDate,
		EndDate:     endDate,
		UsageByType: make(map[string]*tenant.ResourceUsageAnalytics),
		Trends:      make(map[string][]*tenant.UsageTrend),
		Peaks:       make(map[string]*tenant.UsagePeak),
	}

	return analytics, nil
}

// GetPerformanceAnalytics returns performance analytics for a tenant
func (s *TenantServiceImpl) GetPerformanceAnalytics(ctx context.Context, tenantID string, timeRange string) (*tenant.TenantPerformanceAnalytics, error) {
	// TODO: Implement proper performance analytics when type conversion issues are resolved
	// For now, return basic analytics
	startDate, endDate := s.parseTimeRange(timeRange)
	
	analytics := &tenant.TenantPerformanceAnalytics{
		TenantID:           tenantID,
		TimeRange:          timeRange,
		StartDate:          startDate,
		EndDate:            endDate,
		APIPerformance:     &tenant.APIPerformance{},
		ResourcePerformance: make(map[string]*tenant.ResourcePerformance),
		Errors:             []*tenant.ErrorEvent{},
		SlowQueries:        []*tenant.SlowQuery{},
	}

	return analytics, nil
}

// Helper functions
func (s *TenantServiceImpl) parseTimeRange(timeRange string) (time.Time, time.Time) {
	now := time.Now()

	switch timeRange {
	case "1h":
		return now.Add(-1 * time.Hour), now
	case "24h":
		return now.Add(-24 * time.Hour), now
	case "7d":
		return now.Add(-7 * 24 * time.Hour), now
	case "30d":
		return now.Add(-30 * 24 * time.Hour), now
	case "90d":
		return now.Add(-90 * 24 * time.Hour), now
	default:
		return now.Add(-24 * time.Hour), now
	}
}

func (s *TenantServiceImpl) parseAPIRequestEvent(event *tenant.TenantEvent) *tenant.APIRequestEvent {
	// Implementation depends on event structure
	return &tenant.APIRequestEvent{
		Timestamp:    event.Timestamp,
		Endpoint:     "",
		Method:       "",
		StatusCode:   200,
		ResponseTime: 0,
		UserID:       event.UserID,
	}
}

func (s *TenantServiceImpl) parseErrorEvent(event *tenant.TenantEvent) *tenant.ErrorEvent {
	// Implementation depends on event structure
	return &tenant.ErrorEvent{
		Timestamp: event.Timestamp,
		Error:     "",
		Context:   event.EventData,
		UserID:    event.UserID,
	}
}

func (s *TenantServiceImpl) parseSlowQueryEvent(event *tenant.TenantEvent) *tenant.SlowQuery {
	// Implementation depends on event structure
	return &tenant.SlowQuery{
		Timestamp: event.Timestamp,
		Query:     "",
		Duration:  0,
		Context:   event.EventData,
	}
}

func (s *TenantServiceImpl) calculateAPIPerformance(requests []*tenant.APIRequestEvent) *tenant.APIPerformance {
	// Implementation would calculate performance metrics
	return &tenant.APIPerformance{
		TotalRequests:       len(requests),
		AverageResponseTime: 0,
		P95ResponseTime:     0,
		ErrorRate:           0,
		RequestsPerSecond:   0,
	}
}

package services

import (
	"context"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// parseTimeRange parses time range string and returns start and end dates
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

// buildQuotaStatus builds quota status from usage stats
func (s *TenantServiceImpl) buildQuotaStatus(usageStats *tenant.TenantUsageStats) []*tenant.TenantUsage {
	quotaStatus := make([]*tenant.TenantUsage, 0)

	// Create quota status for common resource types
	resourceTypes := []string{"users", "profiles", "carriers", "api_calls", "storage"}

	for _, resourceType := range resourceTypes {
		// Get actual usage data from repository
		usage, err := s.repository.GetUsage(context.Background(), usageStats.TenantID, resourceType)
		if err != nil {
			// Create empty usage if not found
			usage = &tenant.TenantUsage{
				ID:             id.GenerateUsageID(resourceType, usageStats.TenantID),
				TenantID:       usageStats.TenantID,
				ResourceType:   resourceType,
				ResourceCount:  0,
				QuotaLimit:     0,
				QuotaUsed:      0,
				QuotaRemaining: 0,
				PeriodStart:    time.Now().AddDate(0, -1, 0),
				PeriodEnd:      time.Now(),
			}
		}

		quotaStatus = append(quotaStatus, usage)
	}

	return quotaStatus
}

// calculateAPIPerformance calculates real API performance metrics
func (s *TenantServiceImpl) calculateAPIPerformance(requests []*tenant.APIRequestEvent) *tenant.APIPerformance {
	if len(requests) == 0 {
		return &tenant.APIPerformance{
			TotalRequests:       0,
			AverageResponseTime: 0,
			P95ResponseTime:     0,
			ErrorRate:           0,
			RequestsPerSecond:   0,
		}
	}

	// Calculate performance metrics
	totalRequests := len(requests)
	var totalResponseTime int64
	errorCount := 0

	responseTimes := make([]int, 0, totalRequests)

	for _, req := range requests {
		totalResponseTime += int64(req.ResponseTime)
		responseTimes = append(responseTimes, req.ResponseTime)

		if req.StatusCode >= 400 {
			errorCount++
		}
	}

	// Calculate average response time
	averageResponseTime := float64(totalResponseTime) / float64(totalRequests)

	// Calculate P95 response time
	sortedResponseTimes := make([]int, len(responseTimes))
	copy(sortedResponseTimes, responseTimes)

	// Sort response times
	for i := range sortedResponseTimes {
		for j := 0; j < len(sortedResponseTimes)-1-i; j++ {
			if sortedResponseTimes[j] > sortedResponseTimes[j+1] {
				sortedResponseTimes[j], sortedResponseTimes[j+1] = sortedResponseTimes[j+1], sortedResponseTimes[j]
			}
		}
	}

	p95Index := int(float64(totalRequests) * 0.95)
	if p95Index >= totalRequests {
		p95Index = totalRequests - 1
	}
	p95ResponseTime := float64(sortedResponseTimes[p95Index])

	// Calculate error rate
	errorRate := float64(errorCount) / float64(totalRequests) * 100

	// Calculate requests per second
	requestsPerSecond := float64(totalRequests) / 3600.0

	return &tenant.APIPerformance{
		TotalRequests:       totalRequests,
		AverageResponseTime: averageResponseTime,
		P95ResponseTime:     p95ResponseTime,
		ErrorRate:           errorRate,
		RequestsPerSecond:   requestsPerSecond,
	}
}

// parseAPIRequestEvent parses API request event data
func (s *TenantServiceImpl) parseAPIRequestEvent(event *tenant.TenantEvent) *tenant.APIRequestEvent {
	endpoint := ""
	method := "GET"
	responseTime := 0

	// Extract data from event
	if eventData, ok := event.EventData["endpoint"]; ok {
		if ep, ok := eventData.(string); ok {
			endpoint = ep
		}
	}

	if eventData, ok := event.EventData["method"]; ok {
		if m, ok := eventData.(string); ok {
			method = m
		}
	}

	if eventData, ok := event.EventData["response_time"]; ok {
		if rt, ok := eventData.(float64); ok {
			responseTime = int(rt)
		}
	}

	statusCode := 200
	if eventData, ok := event.EventData["status_code"]; ok {
		if sc, ok := eventData.(float64); ok {
			statusCode = int(sc)
		}
	}

	return &tenant.APIRequestEvent{
		Timestamp:    event.Timestamp,
		Endpoint:     endpoint,
		Method:       method,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		UserID:       event.UserID,
	}
}

// parseErrorEvent parses error event data
func (s *TenantServiceImpl) parseErrorEvent(event *tenant.TenantEvent) *tenant.ErrorEvent {
	errorMsg := ""

	if eventData, ok := event.EventData["error"]; ok {
		if err, ok := eventData.(string); ok {
			errorMsg = err
		}
	}

	return &tenant.ErrorEvent{
		Timestamp: event.Timestamp,
		Error:     errorMsg,
		Context:   event.EventData,
		UserID:    event.UserID,
	}
}

// parseSlowQueryEvent parses slow query event data
func (s *TenantServiceImpl) parseSlowQueryEvent(event *tenant.TenantEvent) *tenant.SlowQuery {
	query := ""
	duration := time.Duration(0)

	if eventData, ok := event.EventData["query"]; ok {
		if q, ok := eventData.(string); ok {
			query = q
		}
	}

	if eventData, ok := event.EventData["duration"]; ok {
		if d, ok := eventData.(float64); ok {
			duration = time.Duration(d) * time.Millisecond
		}
	}

	return &tenant.SlowQuery{
		Timestamp: event.Timestamp,
		Query:     query,
		Duration:  duration,
		Context:   event.EventData,
	}
}

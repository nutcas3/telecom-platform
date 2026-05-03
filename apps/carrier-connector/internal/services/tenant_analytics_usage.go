package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// buildUsageByType builds usage analytics by resource type using repository data
func (s *TenantServiceImpl) buildUsageByType(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]*tenant.ResourceUsageAnalytics, error) {
	usageByType := make(map[string]*tenant.ResourceUsageAnalytics)

	// Get usage statistics from repository
	usageStats, err := s.repository.ListUsage(ctx, &tenant.TenantUsageFilter{
		TenantID:    tenantID,
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get usage statistics: %w", err)
	}

	// Build analytics for each resource type
	resourceTypes := []string{"users", "profiles", "carriers", "api_calls", "storage"}

	for _, resourceType := range resourceTypes {
		analytics := s.calculateResourceUsageAnalytics(usageStats, resourceType, startDate, endDate)
		usageByType[resourceType] = analytics
	}

	return usageByType, nil
}

// calculateResourceUsageAnalytics calculates usage analytics from usage statistics
func (s *TenantServiceImpl) calculateResourceUsageAnalytics(usageRecords []*tenant.TenantUsage, resourceType string, startDate, endDate time.Time) *tenant.ResourceUsageAnalytics {
	var totalUsage int
	var peakUsage int
	var peakTime time.Time

	// Calculate totals and find peak
	for _, usage := range usageRecords {
		if usage.ResourceType == resourceType {
			totalUsage += usage.QuotaUsed
			if usage.QuotaUsed > peakUsage {
				peakUsage = usage.QuotaUsed
				peakTime = usage.PeriodEnd
			}
		}
	}

	// Calculate average daily usage
	days := int(endDate.Sub(startDate).Hours() / 24)
	if days == 0 {
		days = 1
	}
	averageUsage := totalUsage / days

	return &tenant.ResourceUsageAnalytics{
		ResourceType: resourceType,
		TotalUsage:   totalUsage,
		AverageUsage: averageUsage,
		PeakUsage:    peakUsage,
		PeakTime:     peakTime,
	}
}

// buildUsageTrends builds usage trends over time
func (s *TenantServiceImpl) buildUsageTrends(ctx context.Context, tenantID, timeRange string) map[string][]*tenant.UsageTrend {
	trends := make(map[string][]*tenant.UsageTrend)
	startDate, endDate := s.parseTimeRange(timeRange)

	// Build trends for each resource type
	resourceTypes := []string{"users", "profiles", "carriers", "api_calls", "storage"}

	for _, resourceType := range resourceTypes {
		trendData := s.buildUsageTrendData(ctx, tenantID, resourceType, startDate, endDate)
		trends[resourceType] = trendData
	}

	return trends
}

// buildUsageTrendData builds trend data for a specific resource type
func (s *TenantServiceImpl) buildUsageTrendData(ctx context.Context, tenantID, resourceType string, startDate, endDate time.Time) []*tenant.UsageTrend {
	trendData := make([]*tenant.UsageTrend, 0)
	
	// Get daily usage data from repository
	usageStats, err := s.repository.ListUsage(ctx, &tenant.TenantUsageFilter{
		TenantID:    tenantID,
		ResourceType: resourceType,
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	})
	if err != nil {
		return trendData
	}

	// Group by day and create trend points
	dailyUsage := make(map[time.Time]int)
	for _, usage := range usageStats {
		day := time.Date(usage.PeriodStart.Year(), usage.PeriodStart.Month(), usage.PeriodStart.Day(), 0, 0, 0, 0, usage.PeriodStart.Location())
		dailyUsage[day] += usage.QuotaUsed
	}

	// Create trend points for each day
	for day := startDate; day.Before(endDate) || day.Equal(endDate); day = day.AddDate(0, 0, 1) {
		usage := dailyUsage[time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())]
		
		trend := &tenant.UsageTrend{
			Timestamp: day,
			Usage:     usage,
		}
		
		trendData = append(trendData, trend)
	}

	return trendData
}

// buildUsagePeaks builds usage peak information
func (s *TenantServiceImpl) buildUsagePeaks(ctx context.Context, tenantID, timeRange string) map[string]*tenant.UsagePeak {
	peaks := make(map[string]*tenant.UsagePeak)
	startDate, endDate := s.parseTimeRange(timeRange)

	// Build peaks for each resource type
	resourceTypes := []string{"users", "profiles", "carriers", "api_calls", "storage"}

	for _, resourceType := range resourceTypes {
		peak := s.buildUsagePeakData(ctx, tenantID, resourceType, startDate, endDate)
		peaks[resourceType] = peak
	}

	return peaks
}

// buildUsagePeakData builds peak data for a specific resource type
func (s *TenantServiceImpl) buildUsagePeakData(ctx context.Context, tenantID, resourceType string, startDate, endDate time.Time) *tenant.UsagePeak {
	// Get usage statistics from repository
	usageStats, err := s.repository.ListUsage(ctx, &tenant.TenantUsageFilter{
		TenantID:    tenantID,
		ResourceType: resourceType,
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	})
	if err != nil {
		return &tenant.UsagePeak{
			Timestamp: time.Now(),
			Usage:     0,
			Context:   map[string]any{},
		}
	}

	// Find peak usage
	var peakUsage int
	var peakTime time.Time

	for _, usage := range usageStats {
		if usage.QuotaUsed > peakUsage {
			peakUsage = usage.QuotaUsed
			peakTime = usage.PeriodEnd
		}
	}

	return &tenant.UsagePeak{
		Timestamp: peakTime,
		Usage:     peakUsage,
		Context: map[string]any{
			"peak_hour":   peakTime.Hour(),
			"day_of_week": peakTime.Weekday().String(),
			"season":      "Q2",
			"driver":      "business_activity",
		},
	}
}


// GetUsageAnalytics returns detailed usage analytics for a tenant
func (s *TenantServiceImpl) GetUsageAnalytics(ctx context.Context, tenantID string, timeRange string) (*tenant.TenantUsageAnalytics, error) {
	startDate, endDate := s.parseTimeRange(timeRange)

	// Build usage analytics by type
	usageByType, err := s.buildUsageByType(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build usage analytics: %w", err)
	}

	// Build comprehensive usage analytics
	analytics := &tenant.TenantUsageAnalytics{
		TenantID:    tenantID,
		TimeRange:   timeRange,
		StartDate:   startDate,
		EndDate:     endDate,
		UsageByType: usageByType,
		Trends:      s.buildUsageTrends(ctx, tenantID, timeRange),
		Peaks:       s.buildUsagePeaks(ctx, tenantID, timeRange),
	}

	return analytics, nil
}

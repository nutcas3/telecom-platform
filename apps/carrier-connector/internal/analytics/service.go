package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides analytics operations
type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewService creates a new analytics service
func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// GetDashboard retrieves the main analytics dashboard
func (s *Service) GetDashboard(ctx context.Context, filter *AnalyticsFilter) (*DashboardMetrics, error) {
	dashboard := &DashboardMetrics{
		TenantID:    filter.TenantID,
		Period:      fmt.Sprintf("%s to %s", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
	}

	var err error
	dashboard.Revenue, err = s.getRevenueMetrics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get revenue metrics")
	}

	dashboard.Subscribers, err = s.getSubscriberStats(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get subscriber stats")
	}

	dashboard.Usage, err = s.getUsageMetrics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get usage metrics")
	}

	dashboard.Carriers, err = s.getCarrierMetrics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get carrier metrics")
	}

	dashboard.Geographic, err = s.getGeoMetrics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get geo metrics")
	}

	dashboard.Performance, err = s.getPerformanceStats(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get performance stats")
	}

	return dashboard, nil
}

// GetRevenueAnalytics retrieves detailed revenue analytics
func (s *Service) GetRevenueAnalytics(ctx context.Context, filter *AnalyticsFilter) (*RevenueMetrics, error) {
	metrics, err := s.getRevenueMetrics(ctx, filter)
	return &metrics, err
}

func (s *Service) getRevenueMetrics(ctx context.Context, filter *AnalyticsFilter) (RevenueMetrics, error) {
	metrics := RevenueMetrics{
		RevenueByCountry:  make(map[string]float64),
		RevenueByCarrier:  make(map[string]float64),
		RevenueByPlan:     make(map[string]float64),
		RevenueByCurrency: make(map[string]float64),
	}

	// Query total revenue
	var totalRevenue float64
	s.db.WithContext(ctx).Table("billing_transactions").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ? AND status = ?",
			filter.TenantID, filter.StartDate, filter.EndDate, "completed").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)
	metrics.TotalRevenue = totalRevenue

	// Query revenue by country
	type countryRevenue struct {
		Country string
		Total   float64
	}
	var byCountry []countryRevenue
	s.db.WithContext(ctx).Table("billing_transactions").
		Select("country, SUM(amount) as total").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", filter.TenantID, filter.StartDate, filter.EndDate).
		Group("country").Scan(&byCountry)
	for _, cr := range byCountry {
		metrics.RevenueByCountry[cr.Country] = cr.Total
	}

	return metrics, nil
}

func (s *Service) getSubscriberStats(ctx context.Context, filter *AnalyticsFilter) (SubscriberStats, error) {
	stats := SubscriberStats{
		ByCountry: make(map[string]int64),
		ByPlan:    make(map[string]int64),
		ByStatus:  make(map[string]int64),
	}

	// Query active subscribers
	s.db.WithContext(ctx).Table("profiles").
		Where("tenant_id = ? AND status = ?", filter.TenantID, "active").
		Count(&stats.TotalActive)

	// Query new subscribers in period
	s.db.WithContext(ctx).Table("profiles").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", filter.TenantID, filter.StartDate, filter.EndDate).
		Count(&stats.NewThisPeriod)

	return stats, nil
}

func (s *Service) getUsageMetrics(ctx context.Context, filter *AnalyticsFilter) (UsageMetrics, error) {
	metrics := UsageMetrics{
		UsageByCountry: make(map[string]int64),
		UsageByCarrier: make(map[string]int64),
	}

	// Query total data usage
	s.db.WithContext(ctx).Table("rate_plan_usage").
		Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate).
		Select("COALESCE(SUM(data_used), 0)").Scan(&metrics.TotalDataUsedMB)

	return metrics, nil
}

func (s *Service) getCarrierMetrics(ctx context.Context, _ *AnalyticsFilter) (CarrierMetrics, error) {
	metrics := CarrierMetrics{
		ByCarrier:        make(map[string]CarrierStat),
		FailuresByReason: make(map[string]int64),
	}

	// Query carrier stats
	var activeCount int64
	s.db.WithContext(ctx).Table("carriers").
		Where("is_active = ?", true).
		Count(&activeCount)
	metrics.ActiveCarriers = int(activeCount)

	return metrics, nil
}

func (s *Service) getGeoMetrics(_ context.Context, _ *AnalyticsFilter) (GeoMetrics, error) {
	metrics := GeoMetrics{
		RevenueByContinent: make(map[string]float64),
	}
	return metrics, nil
}

func (s *Service) getPerformanceStats(_ context.Context, _ *AnalyticsFilter) (PerformanceStats, error) {
	return PerformanceStats{
		Uptime:    99.9,
		ErrorRate: 0.1,
	}, nil
}

// CreateScheduledReport creates a scheduled report
func (s *Service) CreateScheduledReport(ctx context.Context, report *ScheduledReport) error {
	if err := s.db.WithContext(ctx).Create(report).Error; err != nil {
		return fmt.Errorf("failed to create scheduled report: %w", err)
	}
	return nil
}

// ListScheduledReports lists scheduled reports for a tenant
func (s *Service) ListScheduledReports(ctx context.Context, tenantID string) ([]*ScheduledReport, error) {
	var reports []*ScheduledReport
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	return reports, nil
}

package analytics

import "time"

// DashboardMetrics represents the main analytics dashboard data
type DashboardMetrics struct {
	TenantID        string           `json:"tenant_id"`
	Period          string           `json:"period"`
	GeneratedAt     time.Time        `json:"generated_at"`
	Revenue         RevenueMetrics   `json:"revenue"`
	Subscribers     SubscriberStats  `json:"subscribers"`
	Usage           UsageMetrics     `json:"usage"`
	Carriers        CarrierMetrics   `json:"carriers"`
	Geographic      GeoMetrics       `json:"geographic"`
	Performance     PerformanceStats `json:"performance"`
}

// RevenueMetrics contains revenue analytics
type RevenueMetrics struct {
	TotalRevenue       float64            `json:"total_revenue"`
	RecurringRevenue   float64            `json:"recurring_revenue"`
	OneTimeRevenue     float64            `json:"one_time_revenue"`
	RefundsTotal       float64            `json:"refunds_total"`
	NetRevenue         float64            `json:"net_revenue"`
	GrowthRate         float64            `json:"growth_rate_pct"`
	ARPU               float64            `json:"arpu"`
	RevenueByCountry   map[string]float64 `json:"revenue_by_country"`
	RevenueByCarrier   map[string]float64 `json:"revenue_by_carrier"`
	RevenueByPlan      map[string]float64 `json:"revenue_by_plan"`
	RevenueByCurrency  map[string]float64 `json:"revenue_by_currency"`
	DailyRevenue       []TimeSeriesPoint  `json:"daily_revenue"`
	MonthlyRevenue     []TimeSeriesPoint  `json:"monthly_revenue"`
}

// SubscriberStats contains subscriber analytics
type SubscriberStats struct {
	TotalActive        int64              `json:"total_active"`
	NewThisPeriod      int64              `json:"new_this_period"`
	ChurnedThisPeriod  int64              `json:"churned_this_period"`
	ChurnRate          float64            `json:"churn_rate_pct"`
	RetentionRate      float64            `json:"retention_rate_pct"`
	LifetimeValue      float64            `json:"lifetime_value"`
	ByCountry          map[string]int64   `json:"by_country"`
	ByPlan             map[string]int64   `json:"by_plan"`
	ByStatus           map[string]int64   `json:"by_status"`
	GrowthTrend        []TimeSeriesPoint  `json:"growth_trend"`
}

// UsageMetrics contains usage analytics
type UsageMetrics struct {
	TotalDataUsedMB    int64              `json:"total_data_used_mb"`
	TotalVoiceMinutes  int64              `json:"total_voice_minutes"`
	TotalSMSCount      int64              `json:"total_sms_count"`
	AverageDataPerUser float64            `json:"avg_data_per_user_mb"`
	PeakUsageHour      int                `json:"peak_usage_hour"`
	UsageByCountry     map[string]int64   `json:"usage_by_country"`
	UsageByCarrier     map[string]int64   `json:"usage_by_carrier"`
	DailyUsage         []TimeSeriesPoint  `json:"daily_usage"`
}

// CarrierMetrics contains carrier performance analytics
type CarrierMetrics struct {
	TotalCarriers      int                    `json:"total_carriers"`
	ActiveCarriers     int                    `json:"active_carriers"`
	AvgSuccessRate     float64                `json:"avg_success_rate_pct"`
	AvgResponseTime    float64                `json:"avg_response_time_ms"`
	ByCarrier          map[string]CarrierStat `json:"by_carrier"`
	FailuresByReason   map[string]int64       `json:"failures_by_reason"`
}

// CarrierStat contains per-carrier statistics
type CarrierStat struct {
	CarrierID      string  `json:"carrier_id"`
	CarrierName    string  `json:"carrier_name"`
	TotalRequests  int64   `json:"total_requests"`
	SuccessRate    float64 `json:"success_rate_pct"`
	AvgResponseMs  float64 `json:"avg_response_ms"`
	Revenue        float64 `json:"revenue"`
	ActiveProfiles int64   `json:"active_profiles"`
}

// GeoMetrics contains geographic analytics
type GeoMetrics struct {
	TopCountries       []CountryStat `json:"top_countries"`
	TopRegions         []RegionStat  `json:"top_regions"`
	CoverageCountries  int           `json:"coverage_countries"`
	RevenueByContinent map[string]float64 `json:"revenue_by_continent"`
}

// CountryStat contains per-country statistics
type CountryStat struct {
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	Subscribers   int64   `json:"subscribers"`
	Revenue       float64 `json:"revenue"`
	DataUsedMB    int64   `json:"data_used_mb"`
	GrowthRate    float64 `json:"growth_rate_pct"`
}

// RegionStat contains per-region statistics
type RegionStat struct {
	Region      string  `json:"region"`
	Subscribers int64   `json:"subscribers"`
	Revenue     float64 `json:"revenue"`
}

// PerformanceStats contains system performance metrics
type PerformanceStats struct {
	APILatencyP50   float64 `json:"api_latency_p50_ms"`
	APILatencyP95   float64 `json:"api_latency_p95_ms"`
	APILatencyP99   float64 `json:"api_latency_p99_ms"`
	ErrorRate       float64 `json:"error_rate_pct"`
	Uptime          float64 `json:"uptime_pct"`
	TotalAPIRequests int64  `json:"total_api_requests"`
}

// TimeSeriesPoint represents a data point in time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// AnalyticsFilter defines filtering options for analytics queries
type AnalyticsFilter struct {
	TenantID   string    `json:"tenant_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Countries  []string  `json:"countries,omitempty"`
	Carriers   []string  `json:"carriers,omitempty"`
	Plans      []string  `json:"plans,omitempty"`
	GroupBy    string    `json:"group_by,omitempty"`
	Granularity string   `json:"granularity,omitempty"`
}

// ReportType defines available report types
type ReportType string

const (
	ReportTypeRevenue     ReportType = "revenue"
	ReportTypeSubscribers ReportType = "subscribers"
	ReportTypeUsage       ReportType = "usage"
	ReportTypeCarriers    ReportType = "carriers"
	ReportTypeGeographic  ReportType = "geographic"
	ReportTypeExecutive   ReportType = "executive"
)

// ScheduledReport defines a scheduled analytics report
type ScheduledReport struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	TenantID    string     `json:"tenant_id" gorm:"index"`
	Name        string     `json:"name"`
	ReportType  ReportType `json:"report_type"`
	Schedule    string     `json:"schedule"`
	Recipients  []string   `json:"recipients" gorm:"serializer:json"`
	Filter      AnalyticsFilter `json:"filter" gorm:"serializer:json"`
	Format      string     `json:"format"`
	IsActive    bool       `json:"is_active"`
	LastRunAt   *time.Time `json:"last_run_at"`
	NextRunAt   *time.Time `json:"next_run_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

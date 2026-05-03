package telecom

import (
	"context"
	"time"
)

// Subscriber represents a telecom subscriber
type Subscriber struct {
	ID             int64     `json:"id"`
	IMSI           string    `json:"imsi"`
	MSISDN         string    `json:"msisdn"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Status         string    `json:"status"`
	PlanID         int64     `json:"plan_id"`
	Balance        float64   `json:"balance"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SubscriberList represents a paginated list of subscribers
type SubscriberList struct {
	Subscribers []Subscriber `json:"subscribers"`
	Total       int64        `json:"total"`
	Page        int32        `json:"page"`
	PageSize    int32        `json:"page_size"`
	HasNext     bool         `json:"has_next"`
	HasPrev     bool         `json:"has_prev"`
}

// CreateSubscriberRequest represents a request to create a subscriber
type CreateSubscriberRequest struct {
	IMSI           string  `json:"imsi"`
	MSISDN         string  `json:"msisdn"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	PlanID         int64   `json:"plan_id"`
	OrganizationID *string `json:"organization_id,omitempty"`
}

// UpdateSubscriberRequest represents a request to update a subscriber
type UpdateSubscriberRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	PlanID    *int64  `json:"plan_id,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// UsageStats represents usage statistics
type UsageStats struct {
	SubscriberID string    `json:"subscriber_id"`
	DataUp       int64     `json:"data_up"`
	DataDown     int64     `json:"data_down"`
	VoiceSeconds int64     `json:"voice_seconds"`
	SMSCount     int64     `json:"sms_count"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Cost         float64   `json:"cost"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	ID            string                 `json:"id"`
	SubscriberID  string                 `json:"subscriber_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	Gateway       string                 `json:"gateway"`
	TransactionID *string                `json:"transaction_id,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	SubscriberID string                 `json:"subscriber_id"`
	Amount       float64                `json:"amount"`
	Currency     string                 `json:"currency"`
	Gateway      string                 `json:"gateway"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SystemStats represents system statistics
type SystemStats struct {
	ActiveSessions   int64   `json:"active_sessions"`
	TotalAccounts    int64   `json:"total_accounts"`
	BlockedUsers     int64   `json:"blocked_users"`
	LowBalanceAlerts int64   `json:"low_balance_alerts"`
	Uptime           float64 `json:"uptime"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
}

// HealthStatus represents system health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks"`
	Uptime    float64                `json:"uptime"`
}

// RatingPlan represents a rating plan
type RatingPlan struct {
	PlanID     string  `json:"plan_id"`
	Name       string  `json:"name"`
	DataRate   float64 `json:"data_rate"`
	VoiceRate  float64 `json:"voice_rate"`
	SMSRate    float64 `json:"sms_rate"`
	MonthlyFee float64 `json:"monthly_fee"`
	DataLimit  int64   `json:"data_limit"`
	VoiceLimit int64   `json:"voice_limit"`
	SMSLimit   int64   `json:"sms_limit"`
}

// UsageEvent represents a usage event
type UsageEvent struct {
	ID           string                 `json:"id"`
	SubscriberID string                 `json:"subscriber_id"`
	UsageType    string                 `json:"usage_type"`
	Amount       int64                  `json:"amount"`
	Cost         float64                `json:"cost"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CurrentSession represents a current active session
type CurrentSession struct {
	SessionID    string    `json:"session_id"`
	StartTime    time.Time `json:"start_time"`
	DataUp       int64     `json:"data_up"`
	DataDown     int64     `json:"data_down"`
	VoiceSeconds int64     `json:"voice_seconds"`
	SMSCount     int64     `json:"sms_count"`
}

// RealTimeUsage represents real-time usage data
type RealTimeUsage struct {
	CurrentSession *CurrentSession  `json:"current_session,omitempty"`
	TodayUsage     map[string]int64 `json:"today_usage,omitempty"`
}

// gRPC Service Interfaces

// SubscriberServiceClient interface for gRPC subscriber service
type SubscriberServiceClient interface {
	GetSubscriber(ctx context.Context, req *GetSubscriberRequest) (*Subscriber, error)
	ListSubscribers(ctx context.Context, req *ListSubscribersRequest) (*SubscriberList, error)
	CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*Subscriber, error)
	UpdateSubscriber(ctx context.Context, req *UpdateSubscriberRequest) (*Subscriber, error)
	DeleteSubscriber(ctx context.Context, req *DeleteSubscriberRequest) (*DeleteSubscriberResponse, error)
}

// GetSubscriberRequest represents a gRPC request to get a subscriber
type GetSubscriberRequest struct {
	Id int64 `json:"id"`
}

// ListSubscribersRequest represents a gRPC request to list subscribers
type ListSubscribersRequest struct {
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// DeleteSubscriberRequest represents a gRPC request to delete a subscriber
type DeleteSubscriberRequest struct {
	Id int64 `json:"id"`
}

// DeleteSubscriberResponse represents a gRPC response for delete subscriber
type DeleteSubscriberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PaymentServiceClient interface for gRPC payment service
type PaymentServiceClient interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentTransaction, error)
	GetPayment(ctx context.Context, req *GetPaymentRequest) (*PaymentTransaction, error)
	ListPayments(ctx context.Context, req *ListPaymentsRequest) (*PaymentTransactionList, error)
}

// GetPaymentRequest represents a gRPC request to get a payment
type GetPaymentRequest struct {
	Id string `json:"id"`
}

// ListPaymentsRequest represents a gRPC request to list payments
type ListPaymentsRequest struct {
	SubscriberID string `json:"subscriber_id,omitempty"`
	Status       string `json:"status,omitempty"`
	Page         int32  `json:"page"`
	PageSize     int32  `json:"page_size"`
}

// PaymentTransactionList represents a list of payment transactions
type PaymentTransactionList struct {
	Transactions []PaymentTransaction `json:"transactions"`
	Total        int64                `json:"total"`
	Page         int32                `json:"page"`
	PageSize     int32                `json:"page_size"`
	HasNext      bool                 `json:"has_next"`
	HasPrev      bool                 `json:"has_prev"`
}

// UsageServiceClient interface for gRPC usage service
type UsageServiceClient interface {
	GetUsageStats(ctx context.Context, req *GetUsageStatsRequest) (*UsageStats, error)
	GetRealTimeUsage(ctx context.Context, req *GetRealTimeUsageRequest) (*RealTimeUsage, error)
	ListUsageEvents(ctx context.Context, req *ListUsageEventsRequest) (*UsageEventList, error)
}

// GetUsageStatsRequest represents a gRPC request to get usage stats
type GetUsageStatsRequest struct {
	SubscriberID int64     `json:"subscriber_id"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
}

// GetRealTimeUsageRequest represents a gRPC request to get real-time usage
type GetRealTimeUsageRequest struct {
	SubscriberID int64 `json:"subscriber_id"`
}

// ListUsageEventsRequest represents a gRPC request to list usage events
type ListUsageEventsRequest struct {
	SubscriberID *int64     `json:"subscriber_id,omitempty"`
	UsageType    string     `json:"usage_type,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Page         int32      `json:"page"`
	PageSize     int32      `json:"page_size"`
}

// UsageEventList represents a list of usage events
type UsageEventList struct {
	Events   []UsageEvent `json:"events"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"page_size"`
	HasNext  bool         `json:"has_next"`
	HasPrev  bool         `json:"has_prev"`
}

// SystemServiceClient interface for gRPC system service
type SystemServiceClient interface {
	GetSystemStats(ctx context.Context, req *GetSystemStatsRequest) (*SystemStats, error)
	GetHealthStatus(ctx context.Context, req *GetHealthStatusRequest) (*HealthStatus, error)
}

// GetSystemStatsRequest represents a gRPC request to get system stats
type GetSystemStatsRequest struct{}

// GetHealthStatusRequest represents a gRPC request to get health status
type GetHealthStatusRequest struct{}

// ChurnRiskLevel represents the risk level of customer churn
type ChurnRiskLevel string

const (
	ChurnRiskLow      ChurnRiskLevel = "low"
	ChurnRiskMedium   ChurnRiskLevel = "medium"
	ChurnRiskHigh     ChurnRiskLevel = "high"
	ChurnRiskCritical ChurnRiskLevel = "critical"
)

// ChurnPrediction represents a churn prediction for a customer
type ChurnPrediction struct {
	ProfileID          string         `json:"profile_id"`
	RiskLevel          ChurnRiskLevel `json:"risk_level"`
	RiskScore          float64        `json:"risk_score"`
	PredictedChurnDate *time.Time     `json:"predicted_churn_date,omitempty"`
	Reasons            []string       `json:"reasons"`
	Recommendations    []string       `json:"recommendations"`
	LastUpdated        time.Time      `json:"last_updated"`
}

// ChurnMetrics represents churn analysis metrics
type ChurnMetrics struct {
	Period             string                   `json:"period"`
	TotalSubscribers   int64                    `json:"total_subscribers"`
	ChurnedSubscribers int64                    `json:"churned_subscribers"`
	ChurnRate          float64                  `json:"churn_rate"`
	MonthlyChurnRate   float64                  `json:"monthly_churn_rate"`
	AnnualChurnRate    float64                  `json:"annual_churn_rate"`
	AverageTenure      float64                  `json:"average_tenure_days"`
	RiskDistribution   map[ChurnRiskLevel]int64 `json:"risk_distribution"`
	GeneratedAt        time.Time                `json:"generated_at"`
}

// FraudType represents different types of fraud
type FraudType string

const (
	FraudTypeAccountTakeover   FraudType = "account_takeover"
	FraudTypeSubscriptionFraud FraudType = "subscription_fraud"
	FraudTypePaymentFraud      FraudType = "payment_fraud"
	FraudTypeUsageAnomaly      FraudType = "usage_anomaly"
	FraudTypeSIMSwap           FraudType = "sim_swap"
)

// FraudSeverity represents the severity of fraud detection
type FraudSeverity string

const (
	FraudSeverityLow      FraudSeverity = "low"
	FraudSeverityMedium   FraudSeverity = "medium"
	FraudSeverityHigh     FraudSeverity = "high"
	FraudSeverityCritical FraudSeverity = "critical"
)

// FraudAlert represents a fraud detection alert
type FraudAlert struct {
	ID          string         `json:"id"`
	Type        FraudType      `json:"type"`
	Severity    FraudSeverity  `json:"severity"`
	ProfileID   string         `json:"profile_id"`
	Description string         `json:"description"`
	RiskScore   float64        `json:"risk_score"`
	Evidence    []string       `json:"evidence"`
	IPAddress   string         `json:"ip_address"`
	Timestamp   time.Time      `json:"timestamp"`
	Status      string         `json:"status"`
	Actions     []string       `json:"actions_taken"`
	Metadata    map[string]any `json:"metadata"`
}

// FraudMetrics represents fraud detection metrics
type FraudMetrics struct {
	Period            string                  `json:"period"`
	TotalAlerts       int64                   `json:"total_alerts"`
	ResolvedAlerts    int64                   `json:"resolved_alerts"`
	FalsePositives    int64                   `json:"false_positives"`
	ResolutionRate    float64                 `json:"resolution_rate_pct"`
	FalsePositiveRate float64                 `json:"false_positive_rate_pct"`
	ByType            map[FraudType]int64     `json:"by_type"`
	BySeverity        map[FraudSeverity]int64 `json:"by_severity"`
	GeneratedAt       time.Time               `json:"generated_at"`
}

// FraudAlertFilter filters fraud alerts
type FraudAlertFilter struct {
	Type     FraudType     `json:"type,omitempty"`
	Severity FraudSeverity `json:"severity,omitempty"`
	Status   string        `json:"status,omitempty"`
	FromDate *time.Time    `json:"from_date,omitempty"`
	ToDate   *time.Time    `json:"to_date,omitempty"`
	Limit    int           `json:"limit,omitempty"`
}

// MarketMetrics represents market penetration analysis
type MarketMetrics struct {
	Period              string                          `json:"period"`
	TotalMarketSize     int64                           `json:"total_market_size"`
	OurSubscribers      int64                           `json:"our_subscribers"`
	MarketShare         float64                         `json:"market_share_pct"`
	GrowthRate          float64                         `json:"growth_rate_pct"`
	ByCountry           map[string]CountryMetrics       `json:"by_country"`
	ByCarrier           map[string]MarketCarrierMetrics `json:"by_carrier"`
	ByDemographic       map[string]DemoMetrics          `json:"by_demographic"`
	CompetitorAnalysis  map[string]CompetitorMetrics    `json:"competitor_analysis"`
	MarketOpportunities []MarketOpportunity             `json:"market_opportunities"`
	GeneratedAt         time.Time                       `json:"generated_at"`
}

// CountryMetrics represents metrics by country
type CountryMetrics struct {
	Country        string  `json:"country"`
	MarketSize     int64   `json:"market_size"`
	OurSubscribers int64   `json:"our_subscribers"`
	MarketShare    float64 `json:"market_share_pct"`
	GrowthRate     float64 `json:"growth_rate_pct"`
	AverageRevenue float64 `json:"average_revenue"`
}

// MarketCarrierMetrics represents metrics by carrier
type MarketCarrierMetrics struct {
	CarrierID      string  `json:"carrier_id"`
	CarrierName    string  `json:"carrier_name"`
	Subscribers    int64   `json:"subscribers"`
	MarketShare    float64 `json:"market_share_pct"`
	AverageRevenue float64 `json:"average_revenue"`
	QualityScore   float64 `json:"quality_score"`
}

// DemoMetrics represents metrics by demographic
type DemoMetrics struct {
	Segment        string  `json:"segment"`
	Subscribers    int64   `json:"subscribers"`
	MarketShare    float64 `json:"market_share_pct"`
	AverageRevenue float64 `json:"average_revenue"`
	GrowthRate     float64 `json:"growth_rate_pct"`
}

// CompetitorMetrics represents competitor analysis
type CompetitorMetrics struct {
	Name         string   `json:"name"`
	MarketShare  float64  `json:"market_share_pct"`
	Subscribers  int64    `json:"subscribers"`
	AveragePrice float64  `json:"average_price"`
	Strengths    []string `json:"strengths"`
	Weaknesses   []string `json:"weaknesses"`
}

// MarketOpportunity represents a market opportunity
type MarketOpportunity struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Description     string   `json:"description"`
	PotentialSize   int64    `json:"potential_size"`
	Confidence      float64  `json:"confidence"`
	RequiredActions []string `json:"required_actions"`
}

// PredictiveMaintenanceMetrics represents infrastructure health metrics
type PredictiveMaintenanceMetrics struct {
	Period              string                      `json:"period"`
	TotalAssets         int64                       `json:"total_assets"`
	HealthyAssets       int64                       `json:"healthy_assets"`
	AtRiskAssets        int64                       `json:"at_risk_assets"`
	CriticalAssets      int64                       `json:"critical_assets"`
	OverallHealthScore  float64                     `json:"overall_health_score"`
	ByAssetType         map[string]AssetTypeMetrics `json:"by_asset_type"`
	PredictedFailures   []PredictedFailure          `json:"predicted_failures"`
	MaintenanceSchedule []MaintenanceTask           `json:"maintenance_schedule"`
	GeneratedAt         time.Time                   `json:"generated_at"`
}

// AssetTypeMetrics represents metrics by asset type
type AssetTypeMetrics struct {
	AssetType   string  `json:"asset_type"`
	Total       int64   `json:"total"`
	Healthy     int64   `json:"healthy"`
	AtRisk      int64   `json:"at_risk"`
	Critical    int64   `json:"critical"`
	HealthScore float64 `json:"health_score"`
}

// PredictedFailure represents a predicted failure
type PredictedFailure struct {
	AssetID            string    `json:"asset_id"`
	AssetType          string    `json:"asset_type"`
	FailureType        string    `json:"failure_type"`
	PredictedDate      time.Time `json:"predicted_date"`
	Confidence         float64   `json:"confidence"`
	RecommendedActions []string  `json:"recommended_actions"`
}

// MaintenanceTask represents a scheduled maintenance task
type MaintenanceTask struct {
	ID                string    `json:"id"`
	AssetID           string    `json:"asset_id"`
	TaskType          string    `json:"task_type"`
	Priority          string    `json:"priority"`
	ScheduledDate     time.Time `json:"scheduled_date"`
	EstimatedDuration int       `json:"estimated_duration_minutes"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
}

// PricingOptimizationResult represents pricing optimization results
type PricingOptimizationResult struct {
	RatePlanID      string    `json:"rate_plan_id"`
	CurrentPrice    float64   `json:"current_price"`
	OptimalPrice    float64   `json:"optimal_price"`
	Strategy        string    `json:"strategy"`
	ExpectedRevenue float64   `json:"expected_revenue"`
	ExpectedDemand  int64     `json:"expected_demand"`
	PriceChange     float64   `json:"price_change_pct"`
	Reasoning       []string  `json:"reasoning"`
	Risks           []string  `json:"risks"`
	Recommendations []string  `json:"recommendations"`
	Confidence      float64   `json:"confidence"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// PricingMetrics represents pricing optimization metrics
type PricingMetrics struct {
	Period                string    `json:"period"`
	TotalRatePlans        int64     `json:"total_rate_plans"`
	OptimizedRatePlans    int64     `json:"optimized_rate_plans"`
	AveragePriceChange    float64   `json:"average_price_change_pct"`
	ExpectedRevenueImpact float64   `json:"expected_revenue_impact_pct"`
	ChurnRateReduction    float64   `json:"churn_rate_reduction_pct"`
	PriceElasticity       float64   `json:"price_elasticity"`
	CompetitiveIndex      float64   `json:"competitive_index"`
	OptimizationROI       float64   `json:"optimization_roi_pct"`
	GeneratedAt           time.Time `json:"generated_at"`
}

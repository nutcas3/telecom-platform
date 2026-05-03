package rateplan

import (
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// SubscriptionFilter defines filtering options for subscription queries
type SubscriptionFilter struct {
	Status        SubscriptionStatus `json:"status,omitempty"`
	RatePlanID    string             `json:"rate_plan_id,omitempty"`
	StartedAfter  *time.Time         `json:"started_after,omitempty"`
	StartedBefore *time.Time         `json:"started_before,omitempty"`
	Limit         int                `json:"limit,omitempty"`
	Offset        int                `json:"offset,omitempty"`
}

// UsageAnalyticsFilter defines filtering options for usage analytics
type UsageAnalyticsFilter struct {
	RatePlanID string    `json:"rate_plan_id,omitempty"`
	CarrierID  string    `json:"carrier_id,omitempty"`
	Region     string    `json:"region,omitempty"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GroupBy    string    `json:"group_by,omitempty"` // "day", "week", "month"
}

// RevenueAnalyticsFilter defines filtering options for revenue analytics
type RevenueAnalyticsFilter struct {
	RatePlanID string    `json:"rate_plan_id,omitempty"`
	CarrierID  string    `json:"carrier_id,omitempty"`
	Region     string    `json:"region,omitempty"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GroupBy    string    `json:"group_by,omitempty"` // "day", "week", "month"
}

// UsageAnalytics contains usage statistics
type UsageAnalytics struct {
	TotalDataUsed  int64               `json:"total_data_used"`
	TotalVoiceUsed int64               `json:"total_voice_used"`
	TotalSMSUsed   int64               `json:"total_sms_used"`
	ActiveUsers    int                 `json:"active_users"`
	AverageUsage   map[string]float64  `json:"average_usage"`
	UsageByPlan    map[string]int64    `json:"usage_by_plan"`
	UsageByRegion  map[string]int64    `json:"usage_by_region"`
	TimelineData   []TimelineDataPoint `json:"timeline_data"`
}

// RevenueAnalytics contains revenue statistics
type RevenueAnalytics struct {
	TotalRevenue     float64             `json:"total_revenue"`
	RevenueByPlan    map[string]float64  `json:"revenue_by_plan"`
	RevenueByCarrier map[string]float64  `json:"revenue_by_carrier"`
	RevenueByRegion  map[string]float64  `json:"revenue_by_region"`
	AverageRevenue   map[string]float64  `json:"average_revenue"`
	TimelineData     []TimelineDataPoint `json:"timeline_data"`
}

// TimelineDataPoint represents a data point in time series
type TimelineDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

type SearchCriteria struct {
	CarrierID string   `json:"carrier_id,omitempty"`
	Region    string   `json:"region,omitempty"`
	PlanType  PlanType `json:"plan_type,omitempty"`
	MinPrice  float64  `json:"min_price,omitempty"`
	MaxPrice  float64  `json:"max_price,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	SortBy    string   `json:"sort_by,omitempty"`
	SortOrder string   `json:"sort_order,omitempty"`
}

type SubscribeRequest struct {
	ProfileID        string         `json:"profile_id"`
	RatePlanID       string         `json:"rate_plan_id"`
	AutoRenew        bool           `json:"auto_renew"`
	AppliedDiscounts []string       `json:"applied_discounts,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type RecordUsageRequest struct {
	ProfileID string `json:"profile_id"`
	DataUsed  int64  `json:"data_used"`  // in MB
	VoiceUsed int64  `json:"voice_used"` // in minutes
	SMSUsed   int64  `json:"sms_used"`   // count
}

type CalculateCostRequest struct {
	RatePlanID       string   `json:"rate_plan_id"`
	DataUsed         int64    `json:"data_used"`  // in MB
	VoiceUsed        int64    `json:"voice_used"` // in minutes
	SMSUsed          int64    `json:"sms_used"`   // count
	AppliedDiscounts []string `json:"applied_discounts,omitempty"`
}

type PriceOptimization struct {
	CurrentPrice     float64   `json:"current_price"`
	MarketAverage    float64   `json:"market_average"`
	RecommendedPrice float64   `json:"recommended_price"`
	PriceDifference  float64   `json:"price_difference"`
	CompetitorCount  int       `json:"competitor_count"`
	OptimizedAt      time.Time `json:"optimized_at"`
}

type CostBreakdown struct {
	RatePlanID     string     `json:"rate_plan_id"`
	Currency       string     `json:"currency"`
	TotalCost      float64    `json:"total_cost"`
	Subtotal       float64    `json:"subtotal"`
	DiscountTotal  float64    `json:"discount_total"`
	BreakdownItems []CostItem `json:"breakdown_items"`
	CalculatedAt   time.Time  `json:"calculated_at"`
}

type CostItem struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type CarrierRatePlanCriteria struct {
	Region      string         `json:"region"`
	PlanType    PlanType       `json:"plan_type"`
	MaxBudget   float64        `json:"max_budget"`
	Urgency     string         `json:"urgency"`
	Preferences map[string]any `json:"preferences,omitempty"`
}

type RecommendationCriteria struct {
	Region         string   `json:"region"`
	PlanType       PlanType `json:"plan_type"`
	MaxBudget      float64  `json:"max_budget"`
	PreferredData  int64    `json:"preferred_data"`
	PreferredVoice int64    `json:"preferred_voice"`
	PreferredSMS   int64    `json:"preferred_sms"`
	MaxResults     int      `json:"max_results"`
}

type CarrierRatePlanResult struct {
	Carrier    *smdp.Carrier `json:"carrier"`
	RatePlan   *RatePlan     `json:"rate_plan"`
	TotalScore float64       `json:"total_score"`
	SelectedAt time.Time     `json:"selected_at"`
}

type RatePlanRecommendation struct {
	RatePlanID     string          `json:"rate_plan_id"`
	RatePlanName   string          `json:"rate_plan_name"`
	CarrierID      string          `json:"carrier_id"`
	CarrierName    string          `json:"carrier_name"`
	Price          float64         `json:"price"`
	Currency       string          `json:"currency"`
	Relevance      float64         `json:"relevance"`
	Features       []PlanFeature   `json:"features"`
	DataAllowance  *DataAllowance  `json:"data_allowance"`
	VoiceAllowance *VoiceAllowance `json:"voice_allowance"`
	SMSAllowance   *SMSAllowance   `json:"sms_allowance"`
	RecommendedAt  time.Time       `json:"recommended_at"`
}

type CarrierRatePlanAnalytics struct {
	CarrierID     string              `json:"carrier_id"`
	CarrierName   string              `json:"carrier_name"`
	Region        string              `json:"region"`
	HealthStatus  string              `json:"health_status"`
	Priority      int                 `json:"priority"`
	TotalPlans    int                 `json:"total_plans"`
	ActivePlans   int                 `json:"active_plans"`
	GeneratedAt   time.Time           `json:"generated_at"`
	PlanAnalytics []RatePlanAnalytics `json:"plan_analytics"`
}

type RatePlanAnalytics struct {
	RatePlanID          string          `json:"rate_plan_id"`
	RatePlanName        string          `json:"rate_plan_name"`
	BasePrice           float64         `json:"base_price"`
	Currency            string          `json:"currency"`
	ActiveSubscriptions int             `json:"active_subscriptions"`
	PlanType            PlanType        `json:"plan_type"`
	BillingCycle        BillingCycle    `json:"billing_cycle"`
	DataAllowance       *DataAllowance  `json:"data_allowance"`
	VoiceAllowance      *VoiceAllowance `json:"voice_allowance"`
	SMSAllowance        *SMSAllowance   `json:"sms_allowance"`
	Features            []PlanFeature   `json:"features"`
}

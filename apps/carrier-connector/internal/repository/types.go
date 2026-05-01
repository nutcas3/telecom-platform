package repository

import (
	"context"
	"time"
)

// Repository defines the interface for rate plan data operations
type Repository interface {
	// Rate Plan operations
	CreateRatePlan(ctx context.Context, plan *RatePlan) error
	GetRatePlan(ctx context.Context, id string) (*RatePlan, error)
	UpdateRatePlan(ctx context.Context, plan *RatePlan) error
	DeleteRatePlan(ctx context.Context, id string) error
	ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error)

	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error)
	UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error)
	ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error)

	// Usage operations
	CreateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetUsage(ctx context.Context, id string) (*RatePlanUsage, error)
	UpdateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetCurrentUsage(ctx context.Context, profileID string) (*RatePlanUsage, error)
	ListUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error)

	// Analytics operations
	GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error)
	GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error)
	GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error)
}

// RatePlanUsage represents usage data for a rate plan
type RatePlanUsage struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	RatePlanID  string    `json:"rate_plan_id" gorm:"index"`
	ProfileID   string    `json:"profile_id" gorm:"index"`
	CycleStart  time.Time `json:"cycle_start"`
	CycleEnd    time.Time `json:"cycle_end"`
	DataUsed    int64     `json:"data_used"`
	VoiceUsed   int64     `json:"voice_used"`
	SMSUsed     int64     `json:"sms_used"`
	LastUpdated time.Time `json:"last_updated"`
}

// TableName returns the table name for RatePlanUsage
func (RatePlanUsage) TableName() string {
	return "rate_plan_usage"
}

// RatePlanSubscription represents a subscription to a rate plan
type RatePlanSubscription struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ProfileID        string                 `json:"profile_id" gorm:"index"`
	RatePlanID       string                 `json:"rate_plan_id" gorm:"index"`
	Status           SubscriptionStatus     `json:"status" gorm:"index"`
	StartedAt        time.Time              `json:"started_at"`
	EndedAt          *time.Time             `json:"ended_at,omitempty"`
	BillingCycle     BillingCycle           `json:"billing_cycle"`
	NextBillingDate  time.Time              `json:"next_billing_date"`
	AutoRenew        bool                   `json:"auto_renew"`
	CurrentCycle     time.Time              `json:"current_cycle"`
	AppliedDiscounts []string               `json:"applied_discounts,omitempty" gorm:"serializer:json"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// TableName returns the table name for RatePlanSubscription
func (RatePlanSubscription) TableName() string {
	return "rate_plan_subscriptions"
}

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
)

// BillingCycle represents the billing cycle type
type BillingCycle string

const (
	BillingCycleDaily     BillingCycle = "daily"
	BillingCycleWeekly    BillingCycle = "weekly"
	BillingCycleMonthly   BillingCycle = "monthly"
	BillingCycleQuarterly BillingCycle = "quarterly"
	BillingCycleYearly    BillingCycle = "yearly"
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

// RatePlan represents a rate plan
type RatePlan struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	CarrierID      string                 `json:"carrier_id" gorm:"index"`
	Region         string                 `json:"region" gorm:"index"`
	PlanType       PlanType               `json:"plan_type"`
	BasePrice      float64                `json:"base_price"`
	Currency       string                 `json:"currency"`
	BillingCycle   BillingCycle           `json:"billing_cycle"`
	DataAllowance  *DataAllowance         `json:"data_allowance,omitempty" gorm:"serializer:json"`
	VoiceAllowance *VoiceAllowance        `json:"voice_allowance,omitempty" gorm:"serializer:json"`
	SMSAllowance   *SMSAllowance          `json:"sms_allowance,omitempty" gorm:"serializer:json"`
	OverageRates   *OverageRates          `json:"overage_rates,omitempty" gorm:"serializer:json"`
	Discounts      []*Discount            `json:"discounts,omitempty" gorm:"serializer:json"`
	ValidFrom      time.Time              `json:"valid_from"`
	ValidTo        *time.Time             `json:"valid_to,omitempty"`
	IsActive       bool                   `json:"is_active"`
	Status         PlanStatus             `json:"status"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TableName returns the table name for RatePlan
func (RatePlan) TableName() string {
	return "rate_plans"
}

// PlanType represents the type of rate plan
type PlanType string

const (
	PlanTypeData   PlanType = "data"
	PlanTypeVoice  PlanType = "voice"
	PlanTypeSMS    PlanType = "sms"
	PlanTypeBundle PlanType = "bundle"
)

// PlanStatus represents the status of a rate plan
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"
	PlanStatusInactive PlanStatus = "inactive"
	PlanStatusDraft    PlanStatus = "draft"
)

// DataAllowance represents data allowance configuration
type DataAllowance struct {
	Amount    int64  `json:"amount"`
	Unit      string `json:"unit"`
	Unlimited bool   `json:"unlimited"`
}

// VoiceAllowance represents voice allowance configuration
type VoiceAllowance struct {
	Minutes   int64 `json:"minutes"`
	Unlimited bool  `json:"unlimited"`
}

// SMSAllowance represents SMS allowance configuration
type SMSAllowance struct {
	Messages  int64 `json:"messages"`
	Unlimited bool  `json:"unlimited"`
}

// OverageRates represents overage rate configuration
type OverageRates struct {
	DataRate  float64 `json:"data_rate"`
	VoiceRate float64 `json:"voice_rate"`
	SMSRate   float64 `json:"sms_rate"`
}

// Discount represents a discount configuration
type Discount struct {
	ID       string       `json:"id"`
	Type     DiscountType `json:"type"`
	Value    float64      `json:"value"`
	IsActive bool         `json:"is_active"`
}

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

// RatePlanFilter defines filtering options for rate plan queries
type RatePlanFilter struct {
	CarrierID string     `json:"carrier_id,omitempty"`
	Region    string     `json:"region,omitempty"`
	PlanType  PlanType   `json:"plan_type,omitempty"`
	Status    PlanStatus `json:"status,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	MinPrice  float64    `json:"min_price,omitempty"`
	MaxPrice  float64    `json:"max_price,omitempty"`
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`
	SortOrder string     `json:"sort_order,omitempty"`
}

// UsageAnalyticsFilter defines filtering options for usage analytics
type UsageAnalyticsFilter struct {
	RatePlanID string    `json:"rate_plan_id,omitempty"`
	CarrierID  string    `json:"carrier_id,omitempty"`
	Region     string    `json:"region,omitempty"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GroupBy    string    `json:"group_by,omitempty"`
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

// RevenueAnalyticsFilter defines filtering options for revenue analytics
type RevenueAnalyticsFilter struct {
	RatePlanID string    `json:"rate_plan_id,omitempty"`
	CarrierID  string    `json:"carrier_id,omitempty"`
	Region     string    `json:"region,omitempty"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	GroupBy    string    `json:"group_by,omitempty"`
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

// SearchCriteria defines search criteria for rate plans
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

// SubscribeRequest represents a request to subscribe to a rate plan
type SubscribeRequest struct {
	ProfileID        string                 `json:"profile_id"`
	RatePlanID       string                 `json:"rate_plan_id"`
	AutoRenew        bool                   `json:"auto_renew"`
	AppliedDiscounts []string               `json:"applied_discounts,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RecordUsageRequest represents a request to record usage
type RecordUsageRequest struct {
	ProfileID string `json:"profile_id"`
	DataUsed  int64  `json:"data_used"`
	VoiceUsed int64  `json:"voice_used"`
	SMSUsed   int64  `json:"sms_used"`
}

// CalculateCostRequest represents a request to calculate cost
type CalculateCostRequest struct {
	RatePlanID       string   `json:"rate_plan_id"`
	DataUsed         int64    `json:"data_used"`
	VoiceUsed        int64    `json:"voice_used"`
	SMSUsed          int64    `json:"sms_used"`
	AppliedDiscounts []string `json:"applied_discounts,omitempty"`
}

// RatePlanCostCalculation represents the result of a cost calculation
type RatePlanCostCalculation struct {
	RatePlanID   string                 `json:"rate_plan_id"`
	BaseCost     float64                `json:"base_cost"`
	OverageCost  float64                `json:"overage_cost"`
	DiscountCost float64                `json:"discount_cost"`
	TotalCost    float64                `json:"total_cost"`
	Currency     string                 `json:"currency"`
	Breakdown    map[string]interface{} `json:"breakdown"`
	CalculatedAt time.Time              `json:"calculated_at"`
}

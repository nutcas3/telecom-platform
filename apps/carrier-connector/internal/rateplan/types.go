package rateplan

import (
	"time"
)

type DataAllowance struct {
	Type       DataAllowanceType `json:"type"`
	Amount     int64             `json:"amount"` // in MB or GB depending on type
	Unit       string            `json:"unit"`   // "MB", "GB", "TB"
	Unlimited  bool              `json:"unlimited"`
	SpeedLimit *int64            `json:"speed_limit,omitempty"` // in Kbps
}

type DataAllowanceType string

const (
	DataAllowanceTypeDaily    DataAllowanceType = "daily"
	DataAllowanceTypeMonthly  DataAllowanceType = "monthly"
	DataAllowanceTypeCycle    DataAllowanceType = "cycle"
	DataAllowanceTypeLifetime DataAllowanceType = "lifetime"
)

type VoiceAllowance struct {
	Type         VoiceAllowanceType `json:"type"`
	Minutes      int64              `json:"minutes"`
	Unlimited    bool               `json:"unlimited"`
	Destinations []string           `json:"destinations,omitempty"` // country codes
}

type VoiceAllowanceType string

const (
	VoiceAllowanceTypeDaily     VoiceAllowanceType = "daily"
	VoiceAllowanceTypeMonthly   VoiceAllowanceType = "monthly"
	VoiceAllowanceTypeCycle     VoiceAllowanceType = "cycle"
	VoiceAllowanceTypeUnlimited VoiceAllowanceType = "unlimited"
)

type SMSAllowance struct {
	Type         SMSAllowanceType `json:"type"`
	Messages     int64            `json:"messages"`
	Unlimited    bool             `json:"unlimited"`
	Destinations []string         `json:"destinations,omitempty"` // country codes
}

type SMSAllowanceType string

const (
	SMSAllowanceTypeDaily     SMSAllowanceType = "daily"
	SMSAllowanceTypeMonthly   SMSAllowanceType = "monthly"
	SMSAllowanceTypeCycle     SMSAllowanceType = "cycle"
	SMSAllowanceTypeUnlimited SMSAllowanceType = "unlimited"
)

type OverageRates struct {
	DataRate  float64 `json:"data_rate"`  // per MB
	VoiceRate float64 `json:"voice_rate"` // per minute
	SMSRate   float64 `json:"sms_rate"`   // per message
	Currency  string  `json:"currency"`
}

type PlanFeature struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Config      map[string]any `json:"config,omitempty"`
}

type EarlyTermination struct {
	Enabled       bool    `json:"enabled"`
	FeeType       string  `json:"fee_type"`       // "fixed", "percentage", "remaining_months"
	FeeAmount     float64 `json:"fee_amount"`     // for fixed fee type
	FeePercentage float64 `json:"fee_percentage"` // for percentage fee type
	MinMonths     int     `json:"min_months"`     // minimum months before termination fee applies
}

type Discount struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       DiscountType `json:"type"`
	Value      float64      `json:"value"`
	ValidFrom  time.Time    `json:"valid_from"`
	ValidTo    *time.Time   `json:"valid_to,omitempty"`
	Conditions string       `json:"conditions,omitempty"`
	IsActive   bool         `json:"is_active"`
}

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypeRecurring  DiscountType = "recurring"
)

// RatePlanUsage tracks actual usage against a rate plan
type RatePlanUsage struct {
	ID          string    `json:"id" db:"id"`
	RatePlanID  string    `json:"rate_plan_id" db:"rate_plan_id"`
	ProfileID   string    `json:"profile_id" db:"profile_id"`
	CycleStart  time.Time `json:"cycle_start" db:"cycle_start"`
	CycleEnd    time.Time `json:"cycle_end" db:"cycle_end"`
	DataUsed    int64     `json:"data_used" db:"data_used"`   // in MB
	VoiceUsed   int64     `json:"voice_used" db:"voice_used"` // in minutes
	SMSUsed     int64     `json:"sms_used" db:"sms_used"`     // count
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// RatePlanSubscription represents an active subscription to a rate plan
type RatePlanSubscription struct {
	ID               string             `json:"id" db:"id"`
	ProfileID        string             `json:"profile_id" db:"profile_id"`
	RatePlanID       string             `json:"rate_plan_id" db:"rate_plan_id"`
	Status           SubscriptionStatus `json:"status" db:"status"`
	StartedAt        time.Time          `json:"started_at" db:"started_at"`
	EndedAt          *time.Time         `json:"ended_at,omitempty" db:"ended_at"`
	BillingCycle     BillingCycle       `json:"billing_cycle" db:"billing_cycle"`
	NextBillingDate  time.Time          `json:"next_billing_date" db:"next_billing_date"`
	AutoRenew        bool               `json:"auto_renew" db:"auto_renew"`
	CurrentCycle     time.Time          `json:"current_cycle" db:"current_cycle"`
	AppliedDiscounts []string           `json:"applied_discounts,omitempty" db:"applied_discounts"`
	Metadata         map[string]any     `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
}

// SubscriptionStatus defines the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusPending   SubscriptionStatus = "pending"
)

// RatePlanFilter defines filtering options for rate plan queries
type RatePlanFilter struct {
	CarrierID string     `json:"carrier_id,omitempty"`
	Region    string     `json:"region,omitempty"`
	PlanType  PlanType   `json:"plan_type,omitempty"`
	Status    PlanStatus `json:"status,omitempty"`
	MinPrice  float64    `json:"min_price,omitempty"`
	MaxPrice  float64    `json:"max_price,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`
	SortOrder string     `json:"sort_order,omitempty"`
}

// RatePlanCostCalculation represents the result of cost calculation
type RatePlanCostCalculation struct {
	RatePlanID   string         `json:"rate_plan_id"`
	BaseCost     float64        `json:"base_cost"`
	OverageCost  float64        `json:"overage_cost"`
	DiscountCost float64        `json:"discount_cost"`
	TotalCost    float64        `json:"total_cost"`
	Currency     string         `json:"currency"`
	Breakdown    map[string]any `json:"breakdown"`
	CalculatedAt time.Time      `json:"calculated_at"`
}

package rateplan

import (
	"time"
)

type RatePlan struct {
	ID                string                 `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`
	CarrierID         string                 `json:"carrier_id" db:"carrier_id"`
	Region            string                 `json:"region" db:"region"`
	PlanType          PlanType               `json:"plan_type" db:"plan_type"`
	Status            PlanStatus             `json:"status" db:"status"`
	BasePrice         float64                `json:"base_price" db:"base_price"`
	Currency          string                 `json:"currency" db:"currency"`
	BillingCycle      BillingCycle           `json:"billing_cycle" db:"billing_cycle"`
	DataAllowance     *DataAllowance         `json:"data_allowance,omitempty" db:"data_allowance"`
	VoiceAllowance    *VoiceAllowance        `json:"voice_allowance,omitempty" db:"voice_allowance"`
	SMSAllowance      *SMSAllowance          `json:"sms_allowance,omitempty" db:"sms_allowance"`
	OverageRates      *OverageRates          `json:"overage_rates,omitempty" db:"overage_rates"`
	Features          []PlanFeature          `json:"features,omitempty" db:"features"`
	ActivationFee     float64                `json:"activation_fee" db:"activation_fee"`
	EarlyTermination  *EarlyTermination     `json:"early_termination,omitempty" db:"early_termination"`
	Discounts         []Discount             `json:"discounts,omitempty" db:"discounts"`
	ValidFrom         time.Time              `json:"valid_from" db:"valid_from"`
	ValidTo           *time.Time             `json:"valid_to,omitempty" db:"valid_to"`
	Priority          int                    `json:"priority" db:"priority"`
	IsActive          bool                   `json:"is_active" db:"is_active"`
	Metadata          map[string]any         `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// PlanType defines the type of rate plan
type PlanType string

const (
	PlanTypePrepaid     PlanType = "prepaid"
	PlanTypePostpaid    PlanType = "postpaid"
	PlanTypeHybrid      PlanType = "hybrid"
	PlanTypePayAsYouGo  PlanType = "pay_as_you_go"
	PlanTypeUnlimited   PlanType = "unlimited"
)

// PlanStatus defines the status of a rate plan
type PlanStatus string

const (
	PlanStatusDraft     PlanStatus = "draft"
	PlanStatusActive    PlanStatus = "active"
	PlanStatusInactive  PlanStatus = "inactive"
	PlanStatusArchived  PlanStatus = "archived"
)

// BillingCycle defines the billing frequency
type BillingCycle string

const (
	BillingCycleDaily     BillingCycle = "daily"
	BillingCycleWeekly    BillingCycle = "weekly"
	BillingCycleMonthly   BillingCycle = "monthly"
	BillingCycleQuarterly BillingCycle = "quarterly"
	BillingCycleYearly    BillingCycle = "yearly"
)

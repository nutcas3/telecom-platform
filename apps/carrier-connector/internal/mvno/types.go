package mvno

import "time"

// MVNO represents a Mobile Virtual Network Operator
type MVNO struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	Name       string     `json:"name" gorm:"not null"`
	BusinessID string     `json:"business_id" gorm:"uniqueIndex;not null"`
	Status     MVNOStatus `json:"status" gorm:"default:'pending'"`
	Plan       MVNOPlan   `json:"plan" gorm:"not null"`
	Config     MVNOConfig `json:"config" gorm:"serializer:json"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// MVNOStatus represents the onboarding status
type MVNOStatus string

const (
	StatusPending    MVNOStatus = "pending"
	StatusReview     MVNOStatus = "review"
	StatusApproved   MVNOStatus = "approved"
	StatusActive     MVNOStatus = "active"
	StatusSuspended  MVNOStatus = "suspended"
	StatusTerminated MVNOStatus = "terminated"
)

// MVNOPlan represents subscription tiers
type MVNOPlan string

const (
	PlanStarter    MVNOPlan = "starter"
	PlanGrowth     MVNOPlan = "growth"
	PlanScale      MVNOPlan = "scale"
	PlanEnterprise MVNOPlan = "enterprise"
)

// MVNOConfig contains MVNO-specific configuration
type MVNOConfig struct {
	MaxSubscribers    int      `json:"max_subscribers"`
	AllowedCountries  []string `json:"allowed_countries"`
	CarrierPool       []string `json:"carrier_pool"`
	CustomBranding    bool     `json:"custom_branding"`
	APIAccess         bool     `json:"api_access"`
	AdvancedAnalytics bool     `json:"advanced_analytics"`
}

// OnboardingRequest represents a new MVNO onboarding request
type OnboardingRequest struct {
	BusinessName     string   `json:"business_name" binding:"required"`
	BusinessID       string   `json:"business_id" binding:"required"`
	ContactEmail     string   `json:"contact_email" binding:"required,email"`
	ContactPhone     string   `json:"contact_phone" binding:"required"`
	Plan             MVNOPlan `json:"plan" binding:"required"`
	EstimatedSubs    int      `json:"estimated_subscribers" binding:"min:1"`
	TargetCountries  []string `json:"target_countries" binding:"required,min=1"`
	UseCase          string   `json:"use_case" binding:"required"`
	TechnicalContact string   `json:"technical_contact"`
}

// OnboardingProgress tracks the onboarding progress
type OnboardingProgress struct {
	MVNOID      string           `json:"mvno_id"`
	Steps       []OnboardingStep `json:"steps"`
	Progress    float64          `json:"progress"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

// OnboardingStep represents individual onboarding steps
type OnboardingStep struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CompletedAt time.Time `json:"completed_at"`
	Error       string    `json:"error,omitempty"`
}

// MVNOFilter defines filtering options for MVNO queries
type MVNOFilter struct {
	Status        MVNOStatus `json:"status,omitempty"`
	Plan          MVNOPlan   `json:"plan,omitempty"`
	BusinessID    string     `json:"business_id,omitempty"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
}

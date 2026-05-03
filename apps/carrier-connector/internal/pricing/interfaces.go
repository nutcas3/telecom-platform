package pricing

import (
	"context"
)

// Repository defines the interface for pricing data access
type Repository interface {
	// Rule operations
	CreateRule(ctx context.Context, rule *PricingRule) error
	GetRule(ctx context.Context, id string) (*PricingRule, error)
	UpdateRule(ctx context.Context, rule *PricingRule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, filter *PricingFilter) ([]*PricingRule, error)
	CountRules(ctx context.Context, filter *PricingFilter) (int, error)
	
	// Rule evaluation
	GetActiveRules(ctx context.Context, tenantID string) ([]*PricingRule, error)
	GetRulesByType(ctx context.Context, tenantID string, ruleType RuleType) ([]*PricingRule, error)
}

// Service defines the interface for pricing business logic
type Service interface {
	// Rule management
	CreateRule(ctx context.Context, rule *PricingRule) (*PricingRule, error)
	GetRule(ctx context.Context, id string) (*PricingRule, error)
	UpdateRule(ctx context.Context, id string, rule *PricingRule) (*PricingRule, error)
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, filter *PricingFilter) ([]*PricingRule, error)
	
	// Pricing calculations
	CalculatePrice(ctx context.Context, context *PricingContext) (*PricingResult, error)
	ApplyRules(ctx context.Context, context *PricingContext, rules []*PricingRule) (*PricingResult, error)
	ValidateRule(ctx context.Context, rule *PricingRule) error
	
	// Analytics
	GetAnalytics(ctx context.Context, tenantID string) (*PricingAnalytics, error)
}

// RuleEngine defines the interface for rule evaluation
type RuleEngine interface {
	EvaluateRule(ctx context.Context, rule *PricingRule, context *PricingContext) (bool, error)
	ApplyRule(ctx context.Context, rule *PricingRule, context *PricingContext, currentPrice float64) (float64, error)
	ValidateConditions(ctx context.Context, conditions RuleConditions, context *PricingContext) (bool, error)
	ExecuteActions(ctx context.Context, actions RuleActions, currentPrice float64) (float64, error)
}

// PricingEventHandler defines the interface for pricing events
type PricingEventHandler interface {
	OnRuleCreated(ctx context.Context, rule *PricingRule) error
	OnRuleUpdated(ctx context.Context, oldRule, newRule *PricingRule) error
	OnRuleDeleted(ctx context.Context, rule *PricingRule) error
	OnPriceCalculated(ctx context.Context, context *PricingContext, result *PricingResult) error
}

// PricingValidator defines the interface for validation
type PricingValidator interface {
	ValidateRule(ctx context.Context, rule *PricingRule) error
	ValidateContext(ctx context.Context, context *PricingContext) error
	ValidateConditions(ctx context.Context, conditions RuleConditions) error
	ValidateActions(ctx context.Context, actions RuleActions) error
}

// PricingCache defines the interface for caching pricing data
type PricingCache interface {
	GetRule(ctx context.Context, id string) (*PricingRule, error)
	SetRule(ctx context.Context, rule *PricingRule) error
	DeleteRule(ctx context.Context, id string) error
	GetActiveRules(ctx context.Context, tenantID string) ([]*PricingRule, error)
	SetActiveRules(ctx context.Context, tenantID string, rules []*PricingRule) error
	InvalidateTenant(ctx context.Context, tenantID string) error
}

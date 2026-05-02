package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
	"github.com/sirupsen/logrus"
)

// PricingIntegration integrates dynamic pricing with rate plans
type PricingIntegration struct {
	pricingService  pricing.Service
	rateplanService rateplan.Service
	logger          *logrus.Logger
}

// NewPricingIntegration creates a new pricing integration service
func NewPricingIntegration(
	pricingService pricing.Service,
	rateplanService rateplan.Service,
	logger *logrus.Logger,
) *PricingIntegration {
	return &PricingIntegration{
		pricingService:  pricingService,
		rateplanService: rateplanService,
		logger:          logger,
	}
}

// CalculateRatePlanPrice calculates the price for a rate plan with dynamic pricing
func (pi *PricingIntegration) CalculateRatePlanPrice(ctx context.Context, tenantID, ratePlanID string, quantity int, customerID string) (*pricing.PricingResult, error) {
	// Get rate plan
	ratePlan, err := pi.rateplanService.GetRatePlan(ctx, ratePlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Create pricing context
	pricingCtx := &pricing.PricingContext{
		TenantID:   tenantID,
		CustomerID: customerID,
		ProductID:  ratePlanID,
		BasePrice:  ratePlan.BasePrice,
		Currency:   ratePlan.Currency,
		Quantity:   quantity,
		Location:   ratePlan.Region,
		Time:       getCurrentTime(),
		Metadata: map[string]any{
			"rate_plan_name": ratePlan.Name,
			"carrier_id":     ratePlan.CarrierID,
			"plan_type":      ratePlan.PlanType,
		},
	}

	// Calculate price with dynamic pricing
	result, err := pi.pricingService.CalculatePrice(ctx, pricingCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	pi.logger.WithFields(logrus.Fields{
		"tenant_id":      tenantID,
		"rate_plan_id":   ratePlanID,
		"customer_id":    customerID,
		"original_price": result.OriginalPrice,
		"final_price":    result.FinalPrice,
		"discount":       result.Discount,
	}).Info("Rate plan price calculated with dynamic pricing")

	return result, nil
}

// ApplyPricingToSubscription applies pricing rules to a subscription
func (pi *PricingIntegration) ApplyPricingToSubscription(ctx context.Context, subscription *rateplan.RatePlanSubscription) (*rateplan.RatePlanSubscription, error) {
	// Get rate plan
	_, err := pi.rateplanService.GetRatePlan(ctx, subscription.RatePlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Calculate pricing
	result, err := pi.CalculateRatePlanPrice(ctx, subscription.ProfileID, subscription.RatePlanID, 1, subscription.ProfileID)
	if err != nil {
		return nil, err
	}

	// Update subscription metadata with pricing information
	if subscription.Metadata == nil {
		subscription.Metadata = make(map[string]any)
	}

	subscription.Metadata["dynamic_pricing"] = map[string]any{
		"original_price": result.OriginalPrice,
		"final_price":    result.FinalPrice,
		"discount":       result.Discount,
		"applied_rules":  result.AppliedRules,
		"calculated_at":  getCurrentTime(),
	}

	return subscription, nil
}

// CreatePricingRuleFromRatePlan creates a pricing rule based on a rate plan
func (pi *PricingIntegration) CreatePricingRuleFromRatePlan(ctx context.Context, tenantID string, ratePlanID string, ruleType pricing.RuleType, conditions pricing.RuleConditions, actions pricing.RuleActions) (*pricing.PricingRule, error) {
	// Get rate plan
	ratePlan, err := pi.rateplanService.GetRatePlan(ctx, ratePlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Create pricing rule
	rule := &pricing.PricingRule{
		ID:          generateRuleID(),
		Name:        fmt.Sprintf("Dynamic pricing for %s", ratePlan.Name),
		Description: fmt.Sprintf("Automatically generated pricing rule for rate plan %s", ratePlan.Name),
		TenantID:    tenantID,
		Type:        ruleType,
		Priority:    100, // High priority for auto-generated rules
		IsActive:    true,
		Conditions:  conditions,
		Actions:     actions,
		Metadata: map[string]any{
			"rate_plan_id":   ratePlanID,
			"rate_plan_name": ratePlan.Name,
			"auto_generated": true,
			"generated_at":   getCurrentTime(),
		},
	}

	// Create rule
	createdRule, err := pi.pricingService.CreateRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing rule: %w", err)
	}

	pi.logger.WithFields(logrus.Fields{
		"tenant_id":    tenantID,
		"rate_plan_id": ratePlanID,
		"rule_id":      createdRule.ID,
		"rule_type":    ruleType,
	}).Info("Pricing rule created from rate plan")

	return createdRule, nil
}

// GetPricingEffectiveness analyzes the effectiveness of pricing rules
func (pi *PricingIntegration) GetPricingEffectiveness(ctx context.Context, tenantID string) (*PricingEffectiveness, error) {
	// Get pricing analytics
	analytics, err := pi.pricingService.GetAnalytics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing analytics: %w", err)
	}

	// Get rate plan analytics (placeholder - would need actual implementation)
	ratePlanAnalytics := &RatePlanPricingAnalytics{
		TotalRatePlans:   0,
		PlansWithPricing: 0,
		AverageDiscount:  0.0,
		TotalSavings:     0.0,
		ConversionRate:   0.0,
	}

	// Calculate effectiveness
	effectiveness := &PricingEffectiveness{
		TotalRules:            analytics.TotalRules,
		ActiveRules:           analytics.ActiveRules,
		TotalRatePlans:        ratePlanAnalytics.TotalRatePlans,
		PlansWithPricing:      ratePlanAnalytics.PlansWithPricing,
		AverageDiscountRate:   analytics.DiscountStats.AverageDiscount,
		TotalSavings:          analytics.DiscountStats.TotalDiscountValue,
		RulesByType:           analytics.RulesByType,
		ConversionImprovement: ratePlanAnalytics.ConversionRate,
		GeneratedAt:           getCurrentTime(),
	}

	return effectiveness, nil
}

// Supporting types

type PricingEffectiveness struct {
	TotalRules            int            `json:"total_rules"`
	ActiveRules           int            `json:"active_rules"`
	TotalRatePlans        int            `json:"total_rate_plans"`
	PlansWithPricing      int            `json:"plans_with_pricing"`
	AverageDiscountRate   float64        `json:"average_discount_rate"`
	TotalSavings          float64        `json:"total_savings"`
	RulesByType           map[string]int `json:"rules_by_type"`
	ConversionImprovement float64        `json:"conversion_improvement"`
	GeneratedAt           interface{}    `json:"generated_at"`
}

type RatePlanPricingAnalytics struct {
	TotalRatePlans   int     `json:"total_rate_plans"`
	PlansWithPricing int     `json:"plans_with_pricing"`
	AverageDiscount  float64 `json:"average_discount"`
	TotalSavings     float64 `json:"total_savings"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// Helper functions

func generateRuleID() string {
	return fmt.Sprintf("rule_%d", getCurrentTime().UnixNano())
}

func getCurrentTime() time.Time {
	return time.Now()
}

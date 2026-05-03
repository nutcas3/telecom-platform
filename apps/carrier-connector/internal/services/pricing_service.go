package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
	"github.com/sirupsen/logrus"
)

type PricingService struct {
	repository pricing.Repository
	engine     pricing.RuleEngine
	validator  pricing.PricingValidator
	logger     *logrus.Logger
}

func NewPricingService(
	repository pricing.Repository,
	engine pricing.RuleEngine,
	validator pricing.PricingValidator,
	logger *logrus.Logger,
) pricing.Service {
	return &PricingService{
		repository: repository,
		engine:     engine,
		validator:  validator,
		logger:     logger,
	}
}

// ValidateRule validates a pricing rule
func (s *PricingService) ValidateRule(ctx context.Context, rule *pricing.PricingRule) error {
	return s.validator.ValidateRule(ctx, rule)
}

// GetAnalytics retrieves pricing analytics for a tenant
func (s *PricingService) GetAnalytics(ctx context.Context, tenantID string) (*pricing.PricingAnalytics, error) {
	// Get all rules for tenant
	allRules, err := s.repository.ListRules(ctx, &pricing.PricingFilter{
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rules for analytics: %w", err)
	}

	// Get active rules
	activeRules, err := s.repository.GetActiveRules(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active rules for analytics: %w", err)
	}

	// Calculate analytics
	analytics := &pricing.PricingAnalytics{
		TotalRules:  len(allRules),
		ActiveRules: len(activeRules),
		RulesByType: make(map[string]int),
		UsageByRule: make(map[string]int64),
		GeneratedAt: id.GetCurrentTime(),
	}

	// Count rules by type
	for _, rule := range allRules {
		ruleType := string(rule.Type)
		analytics.RulesByType[ruleType]++
	}

	// Calculate actual usage statistics from pricing history
	discountStats, err := s.calculateDiscountStatistics(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to calculate discount statistics, using defaults")
		// Fallback to default values if calculation fails
		discountStats = pricing.DiscountStatistics{
			TotalDiscounts:     0,
			AverageDiscount:    0.0,
			LargestDiscount:    0.0,
			SmallestDiscount:   0.0,
			TotalDiscountValue: 0.0,
		}
	}
	analytics.DiscountStats = discountStats

	return analytics, nil
}

// calculateDiscountStatistics calculates actual discount statistics from pricing history
func (s *PricingService) calculateDiscountStatistics(ctx context.Context) (pricing.DiscountStatistics, error) {
	// Retrieve all rules and compute discount statistics from their action values
	filter := &pricing.PricingFilter{}
	allRules, err := s.repository.ListRules(ctx, filter)
	if err != nil {
		return pricing.DiscountStatistics{}, fmt.Errorf("failed to list rules: %w", err)
	}

	// Filter only active rules
	var activeRules []*pricing.PricingRule
	for _, rule := range allRules {
		if rule.IsActive {
			activeRules = append(activeRules, rule)
		}
	}

	// Calculate discount statistics based on active rules
	stats := pricing.DiscountStatistics{
		TotalDiscounts:     0,
		AverageDiscount:    0.0,
		LargestDiscount:    0.0,
		SmallestDiscount:   0.0,
		TotalDiscountValue: 0.0,
	}

	if len(activeRules) == 0 {
		return stats, nil
	}

	var totalDiscountValue float64
	var discountSum float64
	var largestDiscount float64
	var smallestDiscount float64 = -1

	for _, rule := range activeRules {
		discountValue := extractDiscountValue(rule)

		if discountValue > 0 {
			stats.TotalDiscounts++
			totalDiscountValue += discountValue
			discountSum += discountValue

			if discountValue > largestDiscount {
				largestDiscount = discountValue
			}

			if smallestDiscount == -1 || discountValue < smallestDiscount {
				smallestDiscount = discountValue
			}
		}
	}

	// Calculate final statistics
	stats.TotalDiscountValue = totalDiscountValue
	stats.LargestDiscount = largestDiscount
	stats.SmallestDiscount = smallestDiscount

	if stats.TotalDiscounts > 0 {
		stats.AverageDiscount = discountSum / float64(stats.TotalDiscounts)
	}

	s.logger.WithFields(logrus.Fields{
		"total_discounts":      stats.TotalDiscounts,
		"average_discount":     stats.AverageDiscount,
		"total_discount_value": stats.TotalDiscountValue,
	}).Info("Calculated discount statistics from pricing history")

	return stats, nil
}

// extractDiscountValue extracts discount value from rule actions
func extractDiscountValue(rule *pricing.PricingRule) float64 {
	// Extract the actual discount value from the rule's Actions field
	actionValue := rule.Actions.Value
	if actionValue == 0 {
		return 0.0
	}

	switch rule.Actions.AdjustmentType {
	case pricing.AdjustmentTypePercentage:
		// Value represents the percentage discount directly
		return actionValue
	case pricing.AdjustmentTypeFixed:
		// Value represents a fixed monetary discount
		return actionValue
	case pricing.AdjustmentTypeMultiply:
		// Multiplier < 1.0 implies a discount; convert to percentage
		if actionValue < 1.0 {
			return (1.0 - actionValue) * 100
		}
		return 0.0
	case pricing.AdjustmentTypeOverride:
		// Override replaces the price; if NewPrice is set, compute savings
		if rule.Actions.NewPrice != nil && *rule.Actions.NewPrice < actionValue {
			return actionValue - *rule.Actions.NewPrice
		}
		return 0.0
	default:
		return actionValue
	}
}

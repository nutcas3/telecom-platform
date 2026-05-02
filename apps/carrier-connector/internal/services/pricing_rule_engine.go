package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
)

// PricingRuleEngine implements the rule evaluation logic
type PricingRuleEngine struct {
	validator pricing.PricingValidator
}

// NewPricingRuleEngine creates a new pricing rule engine
func NewPricingRuleEngine(validator pricing.PricingValidator) pricing.RuleEngine {
	return &PricingRuleEngine{
		validator: validator,
	}
}

// EvaluateRule checks if a rule should be applied to the given context
func (e *PricingRuleEngine) EvaluateRule(ctx context.Context, rule *pricing.PricingRule, pricingCtx *pricing.PricingContext) (bool, error) {
	if !rule.IsActive {
		return false, nil
	}

	// Validate conditions
	matches, err := e.ValidateConditions(ctx, rule.Conditions, pricingCtx)
	if err != nil {
		return false, fmt.Errorf("failed to validate conditions: %w", err)
	}

	return matches, nil
}

// ApplyRule applies a rule to adjust the price
func (e *PricingRuleEngine) ApplyRule(ctx context.Context, rule *pricing.PricingRule, pricingCtx *pricing.PricingContext, currentPrice float64) (float64, error) {
	adjustedPrice, err := e.ExecuteActions(ctx, rule.Actions, currentPrice)
	if err != nil {
		return currentPrice, fmt.Errorf("failed to execute actions: %w", err)
	}

	return adjustedPrice, nil
}

// ValidateConditions checks if the conditions match the pricing context
func (e *PricingRuleEngine) ValidateConditions(ctx context.Context, conditions pricing.RuleConditions, pricingCtx *pricing.PricingContext) (bool, error) {
	// Check time range
	if conditions.TimeRange != nil {
		if !e.isWithinTimeRange(conditions.TimeRange, pricingCtx.Time) {
			return false, nil
		}
	}

	// Check geography
	if len(conditions.Geography) > 0 {
		if !e.isInGeography(conditions.Geography, pricingCtx.Location) {
			return false, nil
		}
	}

	// Check customer type
	if len(conditions.CustomerType) > 0 {
		customerType, ok := pricingCtx.Metadata["customer_type"].(string)
		if !ok || !e.isInCustomerType(conditions.CustomerType, customerType) {
			return false, nil
		}
	}

	// Check volume
	if conditions.Volume != nil {
		if !e.isWithinVolumeRange(conditions.Volume, pricingCtx.Quantity) {
			return false, nil
		}
	}

	// Check usage pattern
	if conditions.UsagePattern != nil {
		if !e.matchesUsagePattern(conditions.UsagePattern, pricingCtx.Time) {
			return false, nil
		}
	}

	return true, nil
}

// ExecuteActions applies the pricing adjustments
func (e *PricingRuleEngine) ExecuteActions(ctx context.Context, actions pricing.RuleActions, currentPrice float64) (float64, error) {
	switch actions.AdjustmentType {
	case pricing.AdjustmentTypePercentage:
		adjustedPrice := currentPrice * (1 - actions.Value/100)
		if actions.Limit != nil && adjustedPrice < *actions.Limit {
			adjustedPrice = *actions.Limit
		}
		return adjustedPrice, nil

	case pricing.AdjustmentTypeFixed:
		adjustedPrice := currentPrice - actions.Value
		if actions.Limit != nil && adjustedPrice < *actions.Limit {
			adjustedPrice = *actions.Limit
		}
		return adjustedPrice, nil

	case pricing.AdjustmentTypeMultiply:
		adjustedPrice := currentPrice * actions.Value
		if actions.Limit != nil && adjustedPrice > *actions.Limit {
			adjustedPrice = *actions.Limit
		}
		return adjustedPrice, nil

	case pricing.AdjustmentTypeOverride:
		if actions.NewPrice != nil {
			return *actions.NewPrice, nil
		}
		return currentPrice, fmt.Errorf("override action requires new_price")

	default:
		return currentPrice, fmt.Errorf("unknown adjustment type: %s", actions.AdjustmentType)
	}
}

// Helper methods for condition validation

func (e *PricingRuleEngine) isWithinTimeRange(timeRange *pricing.TimeRange, checkTime time.Time) bool {
	if checkTime.Before(timeRange.Start) || checkTime.After(timeRange.End) {
		return false
	}

	if len(timeRange.Days) > 0 {
		day := strings.ToLower(checkTime.Weekday().String())
		for _, allowedDay := range timeRange.Days {
			if strings.ToLower(allowedDay) == day {
				return true
			}
		}
		return false
	}

	return true
}

func (e *PricingRuleEngine) isInGeography(geography []string, location string) bool {
	if location == "" {
		return false
	}

	for _, geo := range geography {
		if strings.EqualFold(geo, location) {
			return true
		}
	}
	return false
}

func (e *PricingRuleEngine) isInCustomerType(customerTypes []string, customerType string) bool {
	for _, ct := range customerTypes {
		if strings.EqualFold(ct, customerType) {
			return true
		}
	}
	return false
}

func (e *PricingRuleEngine) isWithinVolumeRange(volumeRange *pricing.VolumeRange, quantity int) bool {
	return quantity >= volumeRange.Min && quantity <= volumeRange.Max
}

func (e *PricingRuleEngine) matchesUsagePattern(pattern *pricing.UsagePattern, checkTime time.Time) bool {
	hour := checkTime.Hour()
	
	// Check peak hours
	if len(pattern.PeakHours) > 0 {
		for _, peakHour := range pattern.PeakHours {
			if e.isHourInRange(hour, peakHour) {
				return true
			}
		}
	}

	// Check off-peak hours
	if len(pattern.OffPeakHours) > 0 {
		for _, offPeakHour := range pattern.OffPeakHours {
			if e.isHourInRange(hour, offPeakHour) {
				return true
			}
		}
	}

	return false
}

func (e *PricingRuleEngine) isHourInRange(hour int, hourRange string) bool {
	// Parse hour range like "9-17", "18-23", etc.
	parts := strings.Split(hourRange, "-")
	if len(parts) != 2 {
		return false
	}

	start := e.parseHour(parts[0])
	end := e.parseHour(parts[1])
	
	return hour >= start && hour <= end
}

func (e *PricingRuleEngine) parseHour(hourStr string) int {
	hour := 0
	fmt.Sscanf(hourStr, "%d", &hour)
	return hour
}

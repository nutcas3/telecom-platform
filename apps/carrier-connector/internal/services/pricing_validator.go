package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
)

// PricingValidator implements validation for pricing rules and contexts
type PricingValidator struct{}

// NewPricingValidator creates a new pricing validator
func NewPricingValidator() pricing.PricingValidator {
	return &PricingValidator{}
}

// ValidateRule validates a pricing rule
func (v *PricingValidator) ValidateRule(ctx context.Context, rule *pricing.PricingRule) error {
	if rule.ID == "" {
		return errors.New("rule ID is required")
	}

	if rule.Name == "" {
		return errors.New("rule name is required")
	}

	if rule.TenantID == "" {
		return errors.New("tenant ID is required")
	}

	if rule.Type == "" {
		return errors.New("rule type is required")
	}

	// Validate rule type
	validTypes := map[pricing.RuleType]bool{
		pricing.RuleTypePercentageDiscount: true,
		pricing.RuleTypeFixedDiscount:      true,
		pricing.RuleTypeMultiplier:         true,
		pricing.RuleTypeTieredPricing:      true,
		pricing.RuleTypeDynamicPricing:     true,
		pricing.RuleTypeConditionalPricing: true,
	}

	if !validTypes[rule.Type] {
		return fmt.Errorf("invalid rule type: %s", rule.Type)
	}

	// Validate conditions
	if err := v.ValidateConditions(ctx, rule.Conditions); err != nil {
		return fmt.Errorf("invalid conditions: %w", err)
	}

	// Validate actions
	if err := v.ValidateActions(ctx, rule.Actions); err != nil {
		return fmt.Errorf("invalid actions: %w", err)
	}

	return nil
}

// ValidateContext validates a pricing context
func (v *PricingValidator) ValidateContext(ctx context.Context, pricingCtx *pricing.PricingContext) error {
	if pricingCtx.TenantID == "" {
		return errors.New("tenant ID is required")
	}

	if pricingCtx.ProductID == "" {
		return errors.New("product ID is required")
	}

	if pricingCtx.BasePrice < 0 {
		return errors.New("base price cannot be negative")
	}

	if pricingCtx.Currency == "" {
		return errors.New("currency is required")
	}

	if pricingCtx.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if pricingCtx.Time.IsZero() {
		return errors.New("time is required")
	}

	return nil
}

// ValidateConditions validates rule conditions
func (v *PricingValidator) ValidateConditions(ctx context.Context, conditions pricing.RuleConditions) error {
	// Validate time range
	if conditions.TimeRange != nil {
		if conditions.TimeRange.Start.IsZero() || conditions.TimeRange.End.IsZero() {
			return errors.New("time range requires both start and end times")
		}

		if conditions.TimeRange.Start.After(conditions.TimeRange.End) {
			return errors.New("time range start must be before end")
		}

		// Validate days
		validDays := map[string]bool{
			"monday":    true,
			"tuesday":   true,
			"wednesday": true,
			"thursday":  true,
			"friday":    true,
			"saturday":  true,
			"sunday":    true,
		}

		for _, day := range conditions.TimeRange.Days {
			if !validDays[day] {
				return fmt.Errorf("invalid day: %s", day)
			}
		}
	}

	// Validate geography
	for _, geo := range conditions.Geography {
		if geo == "" {
			return errors.New("geography cannot be empty")
		}
	}

	// Validate customer type
	for _, ct := range conditions.CustomerType {
		if ct == "" {
			return errors.New("customer type cannot be empty")
		}
	}

	// Validate volume range
	if conditions.Volume != nil {
		if conditions.Volume.Min < 0 || conditions.Volume.Max < 0 {
			return errors.New("volume range values must be non-negative")
		}

		if conditions.Volume.Min > conditions.Volume.Max {
			return errors.New("volume min cannot be greater than max")
		}
	}

	// Validate usage pattern
	if conditions.UsagePattern != nil {
		if len(conditions.UsagePattern.PeakHours) == 0 && len(conditions.UsagePattern.OffPeakHours) == 0 {
			return errors.New("usage pattern must have either peak or off-peak hours")
		}

		// Validate hour formats
		allHours := append(conditions.UsagePattern.PeakHours, conditions.UsagePattern.OffPeakHours...)
		for _, hour := range allHours {
			if !v.isValidHourRange(hour) {
				return fmt.Errorf("invalid hour range: %s", hour)
			}
		}
	}

	return nil
}

// ValidateActions validates rule actions
func (v *PricingValidator) ValidateActions(ctx context.Context, actions pricing.RuleActions) error {
	if actions.AdjustmentType == "" {
		return errors.New("adjustment type is required")
	}

	// Validate adjustment type
	validTypes := map[pricing.AdjustmentType]bool{
		pricing.AdjustmentTypePercentage: true,
		pricing.AdjustmentTypeFixed:      true,
		pricing.AdjustmentTypeMultiply:   true,
		pricing.AdjustmentTypeOverride:   true,
	}

	if !validTypes[actions.AdjustmentType] {
		return fmt.Errorf("invalid adjustment type: %s", actions.AdjustmentType)
	}

	// Validate value based on adjustment type
	switch actions.AdjustmentType {
	case pricing.AdjustmentTypePercentage:
		if actions.Value < 0 || actions.Value > 100 {
			return errors.New("percentage value must be between 0 and 100")
		}

	case pricing.AdjustmentTypeFixed:
		if actions.Value < 0 {
			return errors.New("fixed discount value must be non-negative")
		}

	case pricing.AdjustmentTypeMultiply:
		if actions.Value <= 0 {
			return errors.New("multiplier value must be positive")
		}

	case pricing.AdjustmentTypeOverride:
		if actions.NewPrice == nil || *actions.NewPrice < 0 {
			return errors.New("override action requires a valid new price")
		}
	}

	// Validate limit
	if actions.Limit != nil && *actions.Limit < 0 {
		return errors.New("limit cannot be negative")
	}

	return nil
}

// Helper methods

func (v *PricingValidator) isValidHourRange(hourRange string) bool {
	// Should be in format "start-end" like "9-17", "18-23", etc.
	if len(hourRange) < 3 {
		return false
	}

	for i, char := range hourRange {
		if i == 0 && !v.isDigit(char) {
			return false
		}
		if i == len(hourRange)-1 && !v.isDigit(char) {
			return false
		}
		if i == 1 && char != '-' {
			return false
		}
		if i > 1 && i < len(hourRange)-1 && !v.isDigit(char) {
			return false
		}
	}

	return true
}

func (v *PricingValidator) isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

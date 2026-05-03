package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
	"github.com/sirupsen/logrus"
)


// CreateRule creates a new pricing rule
func (s *PricingService) CreateRule(ctx context.Context, rule *pricing.PricingRule) (*pricing.PricingRule, error) {
	// Validate rule
	if err := s.validator.ValidateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create rule
	if err := s.repository.CreateRule(ctx, rule); err != nil {
		s.logger.WithError(err).Error("Failed to create pricing rule")
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"tenant_id": rule.TenantID,
		"type":      rule.Type,
	}).Info("Pricing rule created successfully")

	return rule, nil
}

// GetRule retrieves a pricing rule by ID
func (s *PricingService) GetRule(ctx context.Context, id string) (*pricing.PricingRule, error) {
	rule, err := s.repository.GetRule(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("rule_id", id).Error("Failed to get pricing rule")
		return nil, err
	}

	return rule, nil
}

// UpdateRule updates an existing pricing rule
func (s *PricingService) UpdateRule(ctx context.Context, id string, rule *pricing.PricingRule) (*pricing.PricingRule, error) {
	// Validate rule
	if err := s.validator.ValidateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update rule
	if err := s.repository.UpdateRule(ctx, rule); err != nil {
		s.logger.WithError(err).WithField("rule_id", id).Error("Failed to update pricing rule")
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"tenant_id": rule.TenantID,
		"type":      rule.Type,
	}).Info("Pricing rule updated successfully")

	return rule, nil
}

// DeleteRule deletes a pricing rule
func (s *PricingService) DeleteRule(ctx context.Context, id string) error {
	// Get rule for logging
	rule, err := s.repository.GetRule(ctx, id)
	if err != nil {
		return err
	}

	// Delete rule
	if err := s.repository.DeleteRule(ctx, id); err != nil {
		s.logger.WithError(err).WithField("rule_id", id).Error("Failed to delete pricing rule")
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"rule_id":   rule.ID,
		"tenant_id": rule.TenantID,
		"type":      rule.Type,
	}).Info("Pricing rule deleted successfully")

	return nil
}

// ListRules lists pricing rules with filtering
func (s *PricingService) ListRules(ctx context.Context, filter *pricing.PricingFilter) ([]*pricing.PricingRule, error) {
	rules, err := s.repository.ListRules(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list pricing rules")
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	return rules, nil
}

// CalculatePrice calculates the final price based on active rules
func (s *PricingService) CalculatePrice(ctx context.Context, pricingCtx *pricing.PricingContext) (*pricing.PricingResult, error) {
	// Validate context
	if err := s.validator.ValidateContext(ctx, pricingCtx); err != nil {
		return nil, fmt.Errorf("invalid pricing context: %w", err)
	}

	// Get active rules for tenant
	rules, err := s.repository.GetActiveRules(ctx, pricingCtx.TenantID)
	if err != nil {
		s.logger.WithError(err).WithField("tenant_id", pricingCtx.TenantID).Error("Failed to get active rules")
		return nil, fmt.Errorf("failed to get active rules: %w", err)
	}

	// Apply rules to calculate final price
	result, err := s.ApplyRules(ctx, pricingCtx, rules)
	if err != nil {
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id":      pricingCtx.TenantID,
		"product_id":     pricingCtx.ProductID,
		"original_price": result.OriginalPrice,
		"final_price":    result.FinalPrice,
		"rules_applied":  len(result.AppliedRules),
	}).Info("Price calculated successfully")

	return result, nil
}

// ApplyRules applies specific rules to a pricing context
func (s *PricingService) ApplyRules(ctx context.Context, pricingCtx *pricing.PricingContext, rules []*pricing.PricingRule) (*pricing.PricingResult, error) {
	result := &pricing.PricingResult{
		OriginalPrice: pricingCtx.BasePrice,
		AdjustedPrice: pricingCtx.BasePrice,
		FinalPrice:    pricingCtx.BasePrice,
		Currency:      pricingCtx.Currency,
		AppliedRules:  make([]pricing.AppliedRule, 0),
		Metadata:      make(map[string]any),
	}

	currentPrice := pricingCtx.BasePrice

	// Apply rules in priority order
	for _, rule := range rules {
		shouldApply, err := s.engine.EvaluateRule(ctx, rule, pricingCtx)
		if err != nil {
			s.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to evaluate rule")
			continue
		}

		if shouldApply {
			adjustedPrice, err := s.engine.ApplyRule(ctx, rule, pricingCtx, currentPrice)
			if err != nil {
				s.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to apply rule")
				continue
			}

			// Calculate discount amount
			discount := currentPrice - adjustedPrice

			// Update result
			currentPrice = adjustedPrice
			result.AppliedRules = append(result.AppliedRules, pricing.AppliedRule{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Type:       string(rule.Type),
				Adjustment: discount,
			})

			s.logger.WithFields(logrus.Fields{
				"rule_id":    rule.ID,
				"rule_name":  rule.Name,
				"adjustment": discount,
				"new_price":  adjustedPrice,
			}).Debug("Rule applied")
		}
	}

	// Finalize result
	result.FinalPrice = currentPrice
	result.Discount = result.OriginalPrice - result.FinalPrice

	return result, nil
}

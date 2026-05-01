package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// GetPlansInCurrency gets rate plans with prices converted to a specific currency
func (rpci *RatePlanCurrencyIntegrator) GetPlansInCurrency(ctx context.Context, filter *rateplan.RatePlanFilter, targetCurrency string) ([]*rateplan.RatePlan, error) {
	// Get plans using original filter
	plans, err := rpci.ratePlanService.ListRatePlans(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plans: %w", err)
	}

	// Convert prices to target currency
	for _, plan := range plans {
		if plan.Currency != targetCurrency {
			conversion, err := rpci.exchangeService.ConvertAmount(ctx, plan.BasePrice, plan.Currency, targetCurrency)
			if err != nil {
				rpci.logger.WithError(err).WithFields(logrus.Fields{
					"plan_id":       plan.ID,
					"from_currency": plan.Currency,
					"to_currency":   targetCurrency,
				}).Warn("Failed to convert plan price")
				continue
			}

			// Store original price and update with converted price
			if plan.Metadata == nil {
				plan.Metadata = make(map[string]interface{})
			}
			plan.Metadata["original_price"] = plan.BasePrice
			plan.Metadata["original_currency"] = plan.Currency
			plan.Metadata["exchange_rate"] = conversion.ExchangeRate
			plan.BasePrice = conversion.ConvertedAmount
			plan.Currency = targetCurrency
		}
	}

	return plans, nil
}

// UpdatePlanCurrency updates a rate plan's currency and converts prices
func (rpci *RatePlanCurrencyIntegrator) UpdatePlanCurrency(ctx context.Context, planID string, newCurrency string) error {
	// Get the current plan
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get rate plan: %w", err)
	}

	// If currency is the same, no conversion needed
	if plan.Currency == newCurrency {
		return nil
	}

	// Convert all monetary values to new currency
	convertedPrice, err := rpci.exchangeService.ConvertAmount(ctx, plan.BasePrice, plan.Currency, newCurrency)
	if err != nil {
		return fmt.Errorf("failed to convert base price: %w", err)
	}

	// Update plan with new currency and converted prices
	plan.BasePrice = convertedPrice.ConvertedAmount
	plan.Currency = newCurrency

	// Store conversion information in metadata
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["currency_conversion"] = map[string]interface{}{
		"from_currency": plan.Metadata["original_currency"],
		"to_currency":   newCurrency,
		"exchange_rate": convertedPrice.ExchangeRate,
		"converted_at":  time.Now(),
	}

	// Update the plan
	updatedPlan, err := rpci.ratePlanService.UpdateRatePlan(ctx, plan)
	if err != nil {
		return fmt.Errorf("failed to update rate plan: %w", err)
	}

	// Use the updated plan for logging
	plan = updatedPlan

	rpci.logger.WithFields(logrus.Fields{
		"plan_id":       planID,
		"from_currency": plan.Metadata["original_currency"],
		"to_currency":   newCurrency,
		"exchange_rate": convertedPrice.ExchangeRate,
	}).Info("Rate plan currency updated")

	return nil
}

// calculateOverageCost calculates overage costs for usage
func (rpci *RatePlanCurrencyIntegrator) calculateOverageCost(ctx context.Context, plan *rateplan.RatePlan, usage *rateplan.RatePlanUsage) (float64, error) {
	overageCost := 0.0

	// Calculate data overage
	if plan.DataAllowance != nil && usage.DataUsed > plan.DataAllowance.Amount {
		dataOverage := usage.DataUsed - plan.DataAllowance.Amount
		if plan.OverageRates != nil {
			overageCost += float64(dataOverage) * plan.OverageRates.DataRate
		}
	}

	// Calculate voice overage
	if plan.VoiceAllowance != nil && usage.VoiceUsed > plan.VoiceAllowance.Minutes {
		voiceOverage := usage.VoiceUsed - plan.VoiceAllowance.Minutes
		if plan.OverageRates != nil {
			overageCost += float64(voiceOverage) * plan.OverageRates.VoiceRate
		}
	}

	// Calculate SMS overage
	if plan.SMSAllowance != nil && usage.SMSUsed > plan.SMSAllowance.Messages {
		smsOverage := usage.SMSUsed - plan.SMSAllowance.Messages
		if plan.OverageRates != nil {
			overageCost += float64(smsOverage) * plan.OverageRates.SMSRate
		}
	}

	return overageCost, nil
}

// GetCurrencyUsageForPlan gets currency usage statistics for a specific rate plan
func (rpci *RatePlanCurrencyIntegrator) GetCurrencyUsageForPlan(ctx context.Context, planID string) (map[string]int64, error) {
	// Get all subscriptions for this plan
	filter := &rateplan.SubscriptionFilter{
		RatePlanID: planID,
		Status:     rateplan.SubscriptionStatusActive,
	}

	subscriptions, err := rpci.ratePlanService.ListSubscriptions(ctx, "", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Count currencies
	currencyUsage := make(map[string]int64)

	for _, subscription := range subscriptions {
		currency := rpci.baseCurrency // Default to base currency
		if subscription.Metadata != nil {
			if subCurrency, exists := subscription.Metadata["subscription_currency"]; exists {
				if subCurrencyStr, ok := subCurrency.(string); ok {
					currency = subCurrencyStr
				}
			}
		}
		currencyUsage[currency]++
	}

	return currencyUsage, nil
}

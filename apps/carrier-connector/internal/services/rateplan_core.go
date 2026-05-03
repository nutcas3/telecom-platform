package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

type RatePlanCurrencyIntegrator struct {
	billingService  currency.BillingService
	exchangeService currency.ExchangeRateService
	ratePlanService rateplan.Service
	logger          *logrus.Logger
	baseCurrency    string
}

// NewRatePlanCurrencyIntegrator creates a new rate plan currency integrator
func NewRatePlanCurrencyIntegrator(
	billingService currency.BillingService,
	exchangeService currency.ExchangeRateService,
	ratePlanService rateplan.Service,
	logger *logrus.Logger,
	baseCurrency string,
) *RatePlanCurrencyIntegrator {
	return &RatePlanCurrencyIntegrator{
		billingService:  billingService,
		exchangeService: exchangeService,
		ratePlanService: ratePlanService,
		logger:          logger,
		baseCurrency:    baseCurrency,
	}
}

// SubscribeToPlanWithCurrency subscribes to a rate plan with currency conversion
func (rpci *RatePlanCurrencyIntegrator) SubscribeToPlanWithCurrency(ctx context.Context, profileID string, planID string, targetCurrency string) (*rateplan.RatePlanSubscription, error) {
	// Get the rate plan
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Convert price to requested currency if needed
	subscriptionPrice := plan.BasePrice
	exchangeRate := 1.0

	if targetCurrency != plan.Currency {
		conversion, err := rpci.billingService.ConvertAmount(ctx, &currency.CurrencyConversionRequest{
			Amount:       plan.BasePrice,
			FromCurrency: plan.Currency,
			ToCurrency:   targetCurrency,
		})
		if err != nil {
			rpci.logger.WithError(err).Error("Failed to convert rate plan price")
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		subscriptionPrice = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

	// Create subscription request with currency information
	subscribeReq := &rateplan.SubscribeRequest{
		ProfileID:  profileID,
		RatePlanID: planID,
		AutoRenew:  true,
		Metadata: map[string]any{
			"original_currency":     plan.Currency,
			"subscription_currency": targetCurrency,
			"original_price":        plan.BasePrice,
			"subscription_price":    subscriptionPrice,
			"exchange_rate":         exchangeRate,
		},
	}

	createdSubscription, err := rpci.ratePlanService.SubscribeToPlan(ctx, subscribeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Process initial billing
	billingReq := &currency.BillingRequest{
		ProfileID:      profileID,
		SubscriptionID: createdSubscription.ID,
		Amount:         subscriptionPrice,
		Currency:       targetCurrency,
		Description:    fmt.Sprintf("Initial subscription to %s", plan.Name),
		BillingDate:    time.Now(),
	}

	_, err = rpci.billingService.ProcessBilling(ctx, billingReq)
	if err != nil {
		rpci.logger.WithError(err).Error("Failed to process initial billing")
		// Don't fail the subscription if billing fails, but log it
	}

	rpci.logger.WithFields(logrus.Fields{
		"profile_id":      profileID,
		"plan_id":         planID,
		"currency":        targetCurrency,
		"subscription_id": createdSubscription.ID,
	}).Info("Rate plan subscription created with currency support")

	return createdSubscription, nil
}

// CalculatePlanCostInCurrency calculates the cost of a rate plan in a specific currency
func (rpci *RatePlanCurrencyIntegrator) CalculatePlanCostInCurrency(ctx context.Context, planID string, targetCurrency string, usageData *rateplan.RatePlanUsage) (*currency.BillingSummary, error) {
	// Get the rate plan
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Calculate base cost
	baseCost := plan.BasePrice

	// Add overage costs if usage data is provided
	if usageData != nil {
		overageCost, err := rpci.calculateOverageCost(ctx, plan, usageData)
		if err != nil {
			rpci.logger.WithError(err).Warn("Failed to calculate overage cost")
		} else {
			baseCost += overageCost
		}
	}

	// Convert to requested currency
	convertedCost := baseCost
	exchangeRate := 1.0

	if targetCurrency != plan.Currency {
		conversion, err := rpci.exchangeService.ConvertAmount(ctx, baseCost, plan.Currency, targetCurrency)
		if err != nil {
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		convertedCost = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

	// Create billing summary
	summary := &currency.BillingSummary{
		ProfileID:        usageData.ProfileID,
		TotalAmount:      convertedCost,
		Currency:         targetCurrency,
		BaseTotalAmount:  baseCost,
		BaseCurrency:     plan.Currency,
		TransactionCount: 1,
		FromDate:         time.Now().AddDate(0, -1, 0),
		ToDate:           time.Now(),
		Breakdown: map[string]any{
			"plan_id":           planID,
			"plan_name":         plan.Name,
			"base_cost":         plan.BasePrice,
			"overage_cost":      baseCost - plan.BasePrice,
			"exchange_rate":     exchangeRate,
			"original_currency": plan.Currency,
		},
	}

	return summary, nil
}

// calculateOverageCost calculates overage costs for usage
func (rpci *RatePlanCurrencyIntegrator) calculateOverageCost(_ context.Context, plan *rateplan.RatePlan, usage *rateplan.RatePlanUsage) (float64, error) {
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

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// RatePlanCurrencyIntegrator integrates currency system with rate plans
type RatePlanCurrencyIntegrator struct {
	billingService  currency.BillingService
	exchangeService *ExchangeRateService
	ratePlanService rateplan.Service
	logger          *logrus.Logger
	baseCurrency    string
}

// NewRatePlanCurrencyIntegrator creates a new rate plan currency integrator
func NewRatePlanCurrencyIntegrator(
	billingService currency.BillingService,
	exchangeService *ExchangeRateService,
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
func (rpci *RatePlanCurrencyIntegrator) SubscribeToPlanWithCurrency(ctx context.Context, profileID string, planID string, currency string) (*rateplan.RatePlanSubscription, error) {
	// Get the rate plan
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	// Convert price to requested currency if needed
	subscriptionPrice := plan.BasePrice
	exchangeRate := 1.0

	if currency != plan.Currency {
		conversion, err := rpci.exchangeService.ConvertAmount(ctx, plan.BasePrice, plan.Currency, currency)
		if err != nil {
			rpci.logger.WithError(err).Error("Failed to convert rate plan price")
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		subscriptionPrice = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

	// Create subscription with currency information
	subscription := &rateplan.RatePlanSubscription{
		ProfileID:  profileID,
		RatePlanID: planID,
		Status:     rateplan.SubscriptionStatusActive,
		StartedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"original_currency":     plan.Currency,
			"subscription_currency": currency,
			"original_price":        plan.BasePrice,
			"subscription_price":    subscriptionPrice,
			"exchange_rate":         exchangeRate,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create subscription request
	subscribeReq := &rateplan.SubscribeRequest{
		ProfileID:  profileID,
		RatePlanID: planID,
		AutoRenew:  true,
		Metadata:   subscription.Metadata,
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
		Currency:       currency,
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
		"currency":        currency,
		"subscription_id": createdSubscription.ID,
	}).Info("Rate plan subscription created with currency support")

	return createdSubscription, nil
}

// CalculatePlanCostInCurrency calculates the cost of a rate plan in a specific currency
func (rpci *RatePlanCurrencyIntegrator) CalculatePlanCostInCurrency(ctx context.Context, planID string, currency string, usageData *rateplan.RatePlanUsage) (*currency.BillingSummary, error) {
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

	if currency != plan.Currency {
		conversion, err := rpci.exchangeService.ConvertAmount(ctx, baseCost, plan.Currency, currency)
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
		Currency:         currency,
		BaseTotalAmount:  baseCost,
		BaseCurrency:     plan.Currency,
		TransactionCount: 1,
		FromDate:         time.Now().AddDate(0, -1, 0),
		ToDate:           time.Now(),
		Breakdown: map[string]interface{}{
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

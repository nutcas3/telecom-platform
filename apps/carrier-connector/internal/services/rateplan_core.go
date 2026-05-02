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
	exchangeService *currency.ExchangeRateService
	ratePlanService rateplan.Service
	logger          *logrus.Logger
	baseCurrency    string
}

func NewRatePlanCurrencyIntegrator(
	billingService currency.BillingService,
	exchangeService *currency.ExchangeRateService,
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

func (rpci *RatePlanCurrencyIntegrator) SubscribeToPlanWithCurrency(ctx context.Context, profileID string, planID string, targetCurrency string) (*rateplan.RatePlanSubscription, error) {
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	subscriptionPrice := plan.BasePrice
	exchangeRate := 1.0

	if targetCurrency != plan.Currency {
		conversionReq := &currency.CurrencyConversionRequest{
			Amount:       plan.BasePrice,
			FromCurrency: plan.Currency,
			ToCurrency:   targetCurrency,
		}
		conversion, err := rpci.billingService.ConvertAmount(ctx, conversionReq)
		if err != nil {
			rpci.logger.WithError(err).Error("Failed to convert rate plan price")
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		subscriptionPrice = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

	metadata := map[string]any{
		"original_currency":     plan.Currency,
		"subscription_currency": targetCurrency,
		"original_price":        plan.BasePrice,
		"subscription_price":    subscriptionPrice,
		"exchange_rate":         exchangeRate,
	}

	subscribeReq := &rateplan.SubscribeRequest{
		ProfileID:  profileID,
		RatePlanID: planID,
		AutoRenew:  true,
		Metadata:   metadata,
	}

	createdSubscription, err := rpci.ratePlanService.SubscribeToPlan(ctx, subscribeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

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

func (rpci *RatePlanCurrencyIntegrator) CalculatePlanCostInCurrency(ctx context.Context, planID string, targetCurrency string, usageData *rateplan.RatePlanUsage) (*currency.BillingSummary, error) {
	plan, err := rpci.ratePlanService.GetRatePlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	baseCost := plan.BasePrice

	if usageData != nil {
		overageCost, err := rpci.calculateOverageCost(ctx, plan, usageData)
		if err != nil {
			rpci.logger.WithError(err).Warn("Failed to calculate overage cost")
		} else {
			baseCost += overageCost
		}
	}

	convertedCost := baseCost
	exchangeRate := 1.0

	if targetCurrency != plan.Currency {
		conversionReq := &currency.CurrencyConversionRequest{
			Amount:       baseCost,
			FromCurrency: plan.Currency,
			ToCurrency:   targetCurrency,
		}
		conversion, err := rpci.billingService.ConvertAmount(ctx, conversionReq)
		if err != nil {
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		convertedCost = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

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

func (rpci *RatePlanCurrencyIntegrator) calculateOverageCost(ctx context.Context, plan *rateplan.RatePlan, usage *rateplan.RatePlanUsage) (float64, error) {
	// TODO: Use context for timeout/cancellation in overage calculations
	_ = ctx // Suppress unused parameter warning until implementation is complete
	overageCost := 0.0

	if plan.DataAllowance != nil && usage.DataUsed > plan.DataAllowance.Amount {
		dataOverage := usage.DataUsed - plan.DataAllowance.Amount
		if plan.OverageRates != nil {
			overageCost += float64(dataOverage) * plan.OverageRates.DataRate
		}
	}

	if plan.VoiceAllowance != nil && usage.VoiceUsed > plan.VoiceAllowance.Minutes {
		voiceOverage := usage.VoiceUsed - plan.VoiceAllowance.Minutes
		if plan.OverageRates != nil {
			overageCost += float64(voiceOverage) * plan.OverageRates.VoiceRate
		}
	}

	if plan.SMSAllowance != nil && usage.SMSUsed > plan.SMSAllowance.Messages {
		smsOverage := usage.SMSUsed - plan.SMSAllowance.Messages
		if plan.OverageRates != nil {
			overageCost += float64(smsOverage) * plan.OverageRates.SMSRate
		}
	}

	return overageCost, nil
}

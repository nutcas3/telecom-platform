package rateplan

import (
	"context"
	"fmt"
	"time"
)

type PricingEngine struct {
	repo   Repository
	logger Logger
}

type Logger interface {
	WithError(err error) Logger
	WithField(key string, value any) Logger
	Error(msg string)
	Info(msg string)
	Warning(msg string)
}

func NewPricingEngine(repo Repository, logger Logger) *PricingEngine {
	return &PricingEngine{
		repo:   repo,
		logger: logger,
	}
}

func (pe *PricingEngine) ValidateRatePlan(ctx context.Context, plan *RatePlan) error {
	if err := pe.validateBasicFields(plan); err != nil {
		return err
	}

	if err := pe.validateDates(plan); err != nil {
		return err
	}

	if err := pe.validateAllowances(plan); err != nil {
		return err
	}

	if err := pe.validateOverageRates(plan); err != nil {
		return err
	}

	if err := pe.validateDiscounts(plan); err != nil {
		return err
	}

	if err := pe.validateEarlyTermination(plan); err != nil {
		return err
	}

	return nil
}

func (pe *PricingEngine) CalculateOptimalPrice(ctx context.Context, plan *RatePlan) (*PriceOptimization, error) {
	filter := &RatePlanFilter{
		Region:    plan.Region,
		PlanType:  plan.PlanType,
		Status:    PlanStatusActive,
		IsActive:  &[]bool{true}[0],
		Limit:     10,
	}

	similarPlans, err := pe.repo.ListRatePlans(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar plans: %w", err)
	}

	var totalBasePrice float64
	for _, similarPlan := range similarPlans {
		totalBasePrice += similarPlan.BasePrice
	}

	var marketAverage float64
	if len(similarPlans) > 0 {
		marketAverage = totalBasePrice / float64(len(similarPlans))
	}

	recommendedPrice := pe.calculateRecommendedPrice(plan, marketAverage)

	optimization := &PriceOptimization{
		CurrentPrice:     plan.BasePrice,
		MarketAverage:    marketAverage,
		RecommendedPrice: recommendedPrice,
		PriceDifference:  recommendedPrice - plan.BasePrice,
		CompetitorCount:  len(similarPlans),
		OptimizedAt:      time.Now(),
	}

	return optimization, nil
}

func (pe *PricingEngine) ValidateSubscription(ctx context.Context, req *SubscribeRequest) error {
	plan, err := pe.repo.GetRatePlan(ctx, req.RatePlanID)
	if err != nil {
		return fmt.Errorf("rate plan not found: %w", err)
	}

	if !plan.IsActive || plan.Status != PlanStatusActive {
		return fmt.Errorf("rate plan is not available for subscription")
	}

	now := time.Now()
	if now.Before(plan.ValidFrom) {
		return fmt.Errorf("rate plan is not yet available")
	}

	if plan.ValidTo != nil && now.After(*plan.ValidTo) {
		return fmt.Errorf("rate plan has expired")
	}

	if len(req.AppliedDiscounts) > 0 {
		if err := pe.validateSubscriptionDiscounts(plan, req.AppliedDiscounts); err != nil {
			return err
		}
	}

	return nil
}

func (pe *PricingEngine) CalculateSubscriptionCost(ctx context.Context, req *CalculateCostRequest) (*CostBreakdown, error) {
	// Get the rate plan
	plan, err := pe.repo.GetRatePlan(ctx, req.RatePlanID)
	if err != nil {
		return nil, fmt.Errorf("rate plan not found: %w", err)
	}

	breakdown := &CostBreakdown{
		RatePlanID:     req.RatePlanID,
		Currency:       plan.Currency,
		CalculatedAt:   time.Now(),
		BreakdownItems: []CostItem{},
	}

	baseCost := plan.BasePrice
	breakdown.BreakdownItems = append(breakdown.BreakdownItems, CostItem{
		Type:        "base_price",
		Description: "Base subscription cost",
		Amount:      baseCost,
		Currency:    plan.Currency,
	})

	dataOverageCost := pe.calculateDataOverage(plan, req.DataUsed)
	if dataOverageCost > 0 {
		breakdown.BreakdownItems = append(breakdown.BreakdownItems, CostItem{
			Type:        "data_overage",
			Description: "Data usage over allowance",
			Amount:      dataOverageCost,
			Currency:    plan.Currency,
		})
	}

	voiceOverageCost := pe.calculateVoiceOverage(plan, req.VoiceUsed)
	if voiceOverageCost > 0 {
		breakdown.BreakdownItems = append(breakdown.BreakdownItems, CostItem{
			Type:        "voice_overage",
			Description: "Voice usage over allowance",
			Amount:      voiceOverageCost,
			Currency:    plan.Currency,
		})
	}

	smsOverageCost := pe.calculateSMSOverage(plan, req.SMSUsed)
	if smsOverageCost > 0 {
		breakdown.BreakdownItems = append(breakdown.BreakdownItems, CostItem{
			Type:        "sms_overage",
			Description: "SMS usage over allowance",
			Amount:      smsOverageCost,
			Currency:    plan.Currency,
		})
	}

	discountAmount := pe.calculateDiscounts(plan, req.AppliedDiscounts, baseCost)
	if discountAmount > 0 {
		breakdown.BreakdownItems = append(breakdown.BreakdownItems, CostItem{
			Type:        "discount",
			Description: "Applied discounts",
			Amount:      -discountAmount,
			Currency:    plan.Currency,
		})
	}
	totalCost := baseCost + dataOverageCost + voiceOverageCost + smsOverageCost - discountAmount
	breakdown.TotalCost = totalCost
	breakdown.Subtotal = baseCost + dataOverageCost + voiceOverageCost + smsOverageCost
	breakdown.DiscountTotal = discountAmount

	return breakdown, nil
}


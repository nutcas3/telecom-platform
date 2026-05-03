package services

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

func (a *RatePlanAdapter) SubscribeToPlan(ctx context.Context, req *rateplan.SubscribeRequest) (*rateplan.RatePlanSubscription, error) {
	repoReq := &repository.SubscribeRequest{
		ProfileID:  req.ProfileID,
		RatePlanID: req.RatePlanID,
	}
	subscription, err := a.service.SubscribeToPlan(ctx, repoReq)
	if err != nil {
		return nil, err
	}
	return toRatePlanSub(subscription), nil
}

func (a *RatePlanAdapter) GetSubscription(ctx context.Context, id string) (*rateplan.RatePlanSubscription, error) {
	subscription, err := a.service.GetSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	return toRatePlanSub(subscription), nil
}

func (a *RatePlanAdapter) UpdateSubscription(ctx context.Context, subscription *rateplan.RatePlanSubscription) (*rateplan.RatePlanSubscription, error) {
	repoSub := &repository.RatePlanSubscription{
		ID:         subscription.ID,
		ProfileID:  subscription.ProfileID,
		RatePlanID: subscription.RatePlanID,
		Status:     repository.SubscriptionStatus(subscription.Status),
		CreatedAt:  subscription.CreatedAt,
		UpdatedAt:  subscription.UpdatedAt,
	}
	result, err := a.service.UpdateSubscription(ctx, repoSub)
	if err != nil {
		return nil, err
	}
	return toRatePlanSub(result), nil
}

func (a *RatePlanAdapter) CancelSubscription(ctx context.Context, subscriptionID string, reason string) error {
	return a.service.CancelSubscription(ctx, subscriptionID, reason)
}

func (a *RatePlanAdapter) GetActiveSubscription(ctx context.Context, profileID string) (*rateplan.RatePlanSubscription, error) {
	subscription, err := a.service.GetActiveSubscription(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return toRatePlanSub(subscription), nil
}

func (a *RatePlanAdapter) ListSubscriptions(ctx context.Context, profileID string, filter *rateplan.SubscriptionFilter) ([]*rateplan.RatePlanSubscription, error) {
	repoFilter := &repository.SubscriptionFilter{
		Status: repository.SubscriptionStatus(filter.Status),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	subscriptions, err := a.service.ListSubscriptions(ctx, profileID, repoFilter)
	if err != nil {
		return nil, err
	}
	result := make([]*rateplan.RatePlanSubscription, len(subscriptions))
	for i, sub := range subscriptions {
		result[i] = toRatePlanSub(sub)
	}
	return result, nil
}

func (a *RatePlanAdapter) RecordUsage(ctx context.Context, req *rateplan.RecordUsageRequest) (*rateplan.RatePlanUsage, error) {
	repoReq := &repository.RecordUsageRequest{
		ProfileID: req.ProfileID,
		DataUsed:  req.DataUsed,
		VoiceUsed: req.VoiceUsed,
		SMSUsed:   req.SMSUsed,
	}
	usage, err := a.service.RecordUsage(ctx, repoReq)
	if err != nil {
		return nil, err
	}
	return convertRepoUsage(usage), nil
}

func (a *RatePlanAdapter) GetUsage(ctx context.Context, profileID string) (*rateplan.RatePlanUsage, error) {
	usage, err := a.service.GetUsage(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return convertRepoUsage(usage), nil
}

func (a *RatePlanAdapter) GetUsageHistory(ctx context.Context, profileID string, limit int) ([]*rateplan.RatePlanUsage, error) {
	usageHistory, err := a.service.GetUsageHistory(ctx, profileID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*rateplan.RatePlanUsage, len(usageHistory))
	for i, u := range usageHistory {
		result[i] = convertRepoUsage(u)
	}
	return result, nil
}

func (a *RatePlanAdapter) CalculateCost(ctx context.Context, req *rateplan.CalculateCostRequest) (*rateplan.RatePlanCostCalculation, error) {
	repoReq := &repository.CalculateCostRequest{
		RatePlanID:       req.RatePlanID,
		DataUsed:         req.DataUsed,
		VoiceUsed:        req.VoiceUsed,
		SMSUsed:          req.SMSUsed,
		AppliedDiscounts: req.AppliedDiscounts,
	}
	calc, err := a.service.CalculateCost(ctx, repoReq)
	if err != nil {
		return nil, err
	}
	return &rateplan.RatePlanCostCalculation{
		RatePlanID:   calc.RatePlanID,
		BaseCost:     calc.BaseCost,
		OverageCost:  calc.OverageCost,
		DiscountCost: calc.DiscountCost,
		TotalCost:    calc.TotalCost,
		Currency:     calc.Currency,
		Breakdown:    calc.Breakdown,
		CalculatedAt: calc.CalculatedAt,
	}, nil
}

func (a *RatePlanAdapter) GetUsageAnalytics(ctx context.Context, filter *rateplan.UsageAnalyticsFilter) (*rateplan.UsageAnalytics, error) {
	repoFilter := &repository.UsageAnalyticsFilter{
		RatePlanID: filter.RatePlanID,
		CarrierID:  filter.CarrierID,
		Region:     filter.Region,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		GroupBy:    filter.GroupBy,
	}
	analytics, err := a.service.GetUsageAnalytics(ctx, repoFilter)
	if err != nil {
		return nil, err
	}
	return &rateplan.UsageAnalytics{
		TotalDataUsed:  analytics.TotalDataUsed,
		TotalVoiceUsed: analytics.TotalVoiceUsed,
		TotalSMSUsed:   analytics.TotalSMSUsed,
		ActiveUsers:    analytics.ActiveUsers,
		AverageUsage:   analytics.AverageUsage,
		UsageByPlan:    analytics.UsageByPlan,
		UsageByRegion:  analytics.UsageByRegion,
	}, nil
}

func (a *RatePlanAdapter) GetRevenueAnalytics(ctx context.Context, filter *rateplan.RevenueAnalyticsFilter) (*rateplan.RevenueAnalytics, error) {
	repoFilter := &repository.RevenueAnalyticsFilter{
		RatePlanID: filter.RatePlanID,
		CarrierID:  filter.CarrierID,
		Region:     filter.Region,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		GroupBy:    filter.GroupBy,
	}
	analytics, err := a.service.GetRevenueAnalytics(ctx, repoFilter)
	if err != nil {
		return nil, err
	}
	return &rateplan.RevenueAnalytics{
		TotalRevenue:     analytics.TotalRevenue,
		RevenueByPlan:    analytics.RevenueByPlan,
		RevenueByCarrier: analytics.RevenueByCarrier,
		RevenueByRegion:  analytics.RevenueByRegion,
		AverageRevenue:   analytics.AverageRevenue,
	}, nil
}

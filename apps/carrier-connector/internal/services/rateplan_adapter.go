package services

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

// RatePlanAdapter adapts services.Service to implement rateplan.Service interface
type RatePlanAdapter struct {
	service *Service
}

// NewRatePlanAdapter creates a new rate plan adapter
func NewRatePlanAdapter(service *Service) rateplan.Service {
	return &RatePlanAdapter{service: service}
}

func (a *RatePlanAdapter) CreateRatePlan(ctx context.Context, plan *rateplan.RatePlan) (*rateplan.RatePlan, error) {
	result, err := a.service.CreateRatePlan(ctx, toRepoPlan(plan))
	if err != nil {
		return nil, err
	}
	return toRatePlan(result), nil
}

func (a *RatePlanAdapter) GetRatePlan(ctx context.Context, id string) (*rateplan.RatePlan, error) {
	plan, err := a.service.GetRatePlan(ctx, id)
	if err != nil {
		return nil, err
	}
	return toRatePlan(plan), nil
}

func (a *RatePlanAdapter) UpdateRatePlan(ctx context.Context, plan *rateplan.RatePlan) (*rateplan.RatePlan, error) {
	result, err := a.service.UpdateRatePlan(ctx, toRepoPlan(plan))
	if err != nil {
		return nil, err
	}
	return toRatePlan(result), nil
}

func (a *RatePlanAdapter) DeleteRatePlan(ctx context.Context, id string) error {
	return a.service.DeleteRatePlan(ctx, id)
}

func (a *RatePlanAdapter) ListRatePlans(ctx context.Context, filter *rateplan.RatePlanFilter) ([]*rateplan.RatePlan, error) {
	repoFilter := &repository.RatePlanFilter{
		CarrierID: filter.CarrierID,
		Region:    filter.Region,
		PlanType:  repository.PlanType(filter.PlanType),
		IsActive:  filter.IsActive,
		MinPrice:  filter.MinPrice,
		MaxPrice:  filter.MaxPrice,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}
	plans, err := a.service.ListRatePlans(ctx, repoFilter)
	if err != nil {
		return nil, err
	}
	return toRatePlanSlice(plans), nil
}

func (a *RatePlanAdapter) SearchRatePlans(ctx context.Context, criteria rateplan.SearchCriteria) ([]*rateplan.RatePlan, error) {
	filter := &rateplan.RatePlanFilter{
		CarrierID: criteria.CarrierID,
		Region:    criteria.Region,
		PlanType:  criteria.PlanType,
	}
	return a.ListRatePlans(ctx, filter)
}

func (a *RatePlanAdapter) GetPopularPlans(ctx context.Context, limit int) ([]*rateplan.RatePlan, error) {
	plans, err := a.service.GetPopularPlans(ctx, limit)
	if err != nil {
		return nil, err
	}
	return toRatePlanSlice(plans), nil
}

// Conversion helpers

func toRepoPlan(p *rateplan.RatePlan) *repository.RatePlan {
	return &repository.RatePlan{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		CarrierID:    p.CarrierID,
		Region:       p.Region,
		PlanType:     repository.PlanType(p.PlanType),
		BasePrice:    p.BasePrice,
		BillingCycle: repository.BillingCycle(p.BillingCycle),
		ValidFrom:    p.ValidFrom,
		ValidTo:      p.ValidTo,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toRatePlan(p *repository.RatePlan) *rateplan.RatePlan {
	return &rateplan.RatePlan{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		CarrierID:    p.CarrierID,
		Region:       p.Region,
		PlanType:     rateplan.PlanType(p.PlanType),
		BasePrice:    p.BasePrice,
		BillingCycle: rateplan.BillingCycle(p.BillingCycle),
		ValidFrom:    p.ValidFrom,
		ValidTo:      p.ValidTo,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toRatePlanSlice(plans []*repository.RatePlan) []*rateplan.RatePlan {
	result := make([]*rateplan.RatePlan, len(plans))
	for i, p := range plans {
		result[i] = toRatePlan(p)
	}
	return result
}

func toRatePlanSub(s *repository.RatePlanSubscription) *rateplan.RatePlanSubscription {
	return &rateplan.RatePlanSubscription{
		ID:         s.ID,
		ProfileID:  s.ProfileID,
		RatePlanID: s.RatePlanID,
		Status:     rateplan.SubscriptionStatus(s.Status),
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func convertRepoUsage(u *repository.RatePlanUsage) *rateplan.RatePlanUsage {
	return &rateplan.RatePlanUsage{
		ID:          u.ID,
		RatePlanID:  u.RatePlanID,
		ProfileID:   u.ProfileID,
		CycleStart:  u.CycleStart,
		CycleEnd:    u.CycleEnd,
		DataUsed:    u.DataUsed,
		VoiceUsed:   u.VoiceUsed,
		SMSUsed:     u.SMSUsed,
		LastUpdated: u.LastUpdated,
	}
}

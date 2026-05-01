package services

import (
	"fmt"
	"time"
)

func (s *Service) validateRatePlan(plan *RatePlan) error {
	if plan.Name == "" {
		return fmt.Errorf("rate plan name is required")
	}
	if plan.CarrierID == "" {
		return fmt.Errorf("carrier ID is required")
	}
	if plan.Region == "" {
		return fmt.Errorf("region is required")
	}
	if plan.BasePrice < 0 {
		return fmt.Errorf("base price cannot be negative")
	}
	if plan.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if plan.BillingCycle == "" {
		return fmt.Errorf("billing cycle is required")
	}
	if plan.ValidFrom.IsZero() {
		return fmt.Errorf("valid from date is required")
	}
	return nil
}

func (s *Service) validateSubscribeRequest(req *SubscribeRequest) error {
	if req.ProfileID == "" {
		return fmt.Errorf("profile ID is required")
	}
	if req.RatePlanID == "" {
		return fmt.Errorf("rate plan ID is required")
	}
	return nil
}

func (s *Service) calculateNextBillingDate(cycle BillingCycle, from time.Time) time.Time {
	switch cycle {
	case BillingCycleDaily:
		return from.AddDate(0, 0, 1)
	case BillingCycleWeekly:
		return from.AddDate(0, 0, 7)
	case BillingCycleMonthly:
		return from.AddDate(0, 1, 0)
	case BillingCycleQuarterly:
		return from.AddDate(0, 3, 0)
	case BillingCycleYearly:
		return from.AddDate(1, 0, 0)
	default:
		return from.AddDate(0, 1, 0) // Default to monthly
	}
}

func (s *Service) calculateCycleEnd(cycle BillingCycle, cycleStart time.Time) time.Time {
	switch cycle {
	case BillingCycleDaily:
		return cycleStart.AddDate(0, 0, 1).Add(-time.Nanosecond)
	case BillingCycleWeekly:
		return cycleStart.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case BillingCycleMonthly:
		return cycleStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case BillingCycleQuarterly:
		return cycleStart.AddDate(0, 3, 0).Add(-time.Nanosecond)
	case BillingCycleYearly:
		return cycleStart.AddDate(1, 0, 0).Add(-time.Nanosecond)
	default:
		return cycleStart.AddDate(0, 1, 0).Add(-time.Nanosecond) // Default to monthly
	}
}


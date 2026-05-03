package pricing

import (
	"time"
)

// getPeriodStart returns start date for period
func (s *PricingOptimizationService) getPeriodStart(period string) time.Time {
	now := time.Now()
	switch period {
	case "daily":
		return now.Truncate(24 * time.Hour)
	case "weekly":
		return now.AddDate(0, 0, -7)
	case "monthly":
		return now.AddDate(0, -1, 0)
	case "quarterly":
		return now.AddDate(0, -3, 0)
	default:
		return now.AddDate(0, -1, 0)
	}
}

// getPeriodEnd returns end date for period
func (s *PricingOptimizationService) getPeriodEnd(period string) time.Time {
	return time.Now()
}

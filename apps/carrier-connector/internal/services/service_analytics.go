package services

import (
	"context"
)

// GetUsage retrieves usage for a profile
func (s *Service) GetUsage(ctx context.Context, profileID string) (*RatePlanUsage, error) {
	usage, err := s.repo.GetCurrentUsage(ctx, profileID)
	if err != nil {
		s.logger.WithError(err).WithField("profile_id", profileID).Error("Failed to get usage")
		return nil, err
	}

	return usage, nil
}
// GetUsageHistory retrieves usage history for a profile
func (s *Service) GetUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error) {
	usageHistory, err := s.repo.ListUsageHistory(ctx, profileID, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get usage history")
		return nil, err
	}

	return usageHistory, nil
}

// GetUsageAnalytics retrieves usage analytics
func (s *Service) GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error) {
	analytics, err := s.repo.GetUsageAnalytics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get usage analytics")
		return nil, err
	}

	return analytics, nil
}

// GetRevenueAnalytics retrieves revenue analytics
func (s *Service) GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error) {
	analytics, err := s.repo.GetRevenueAnalytics(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get revenue analytics")
		return nil, err
	}

	return analytics, nil
}

// GetPopularPlans retrieves the most popular rate plans
func (s *Service) GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error) {
	plans, err := s.repo.GetPopularPlans(ctx, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get popular plans")
		return nil, err
	}

	return plans, nil
}

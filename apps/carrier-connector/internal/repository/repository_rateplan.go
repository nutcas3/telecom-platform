package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CreateUsage creates a new usage record
func (r *GormRepository) CreateUsage(ctx context.Context, usage *RatePlanUsage) error {
	usage.LastUpdated = time.Now()

	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create usage record")
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	return nil
}

// GetUsage retrieves a usage record by ID
func (r *GormRepository) GetUsage(ctx context.Context, id string) (*RatePlanUsage, error) {
	var usage RatePlanUsage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("usage record not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get usage record")
		return nil, fmt.Errorf("failed to get usage record: %w", err)
	}

	return &usage, nil
}

// UpdateUsage updates an existing usage record
func (r *GormRepository) UpdateUsage(ctx context.Context, usage *RatePlanUsage) error {
	usage.LastUpdated = time.Now()

	result := r.db.WithContext(ctx).Where("id = ?", usage.ID).Updates(usage)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update usage record")
		return fmt.Errorf("failed to update usage record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("usage record not found: %s", usage.ID)
	}

	return nil
}

// GetCurrentUsage retrieves the current usage for a profile
func (r *GormRepository) GetCurrentUsage(ctx context.Context, profileID string) (*RatePlanUsage, error) {
	var usage RatePlanUsage
	err := r.db.WithContext(ctx).
		Where("profile_id = ?", profileID).
		Order("cycle_start DESC").
		First(&usage).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No usage record found
		}
		r.logger.WithError(err).Error("Failed to get current usage")
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return &usage, nil
}

// ListUsageHistory retrieves usage history for a profile
func (r *GormRepository) ListUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error) {
	query := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Order("cycle_start DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var usageHistory []*RatePlanUsage
	if err := query.Find(&usageHistory).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list usage history")
		return nil, fmt.Errorf("failed to list usage history: %w", err)
	}

	return usageHistory, nil
}

// GetUsageAnalytics retrieves usage analytics
func (r *GormRepository) GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error) {
	analytics := &UsageAnalytics{
		TotalDataUsed:  0,
		TotalVoiceUsed: 0,
		TotalSMSUsed:   0,
		ActiveUsers:    0,
		AverageUsage:   make(map[string]float64),
		UsageByPlan:    make(map[string]int64),
		UsageByRegion:  make(map[string]int64),
		TimelineData:   []TimelineDataPoint{},
	}

	// Get active users count
	var activeUsersCount int64
	if err := r.db.WithContext(ctx).
		Model(&RatePlanSubscription{}).
		Where("status = ?", SubscriptionStatusActive).
		Count(&activeUsersCount).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get active users count")
		return nil, fmt.Errorf("failed to get active users count: %w", err)
	}

	// Get total usage
	var totalDataUsed, totalVoiceUsed, totalSMSUsed int64
	if err := r.db.WithContext(ctx).
		Model(&RatePlanUsage{}).
		Select("COALESCE(SUM(data_used), 0), COALESCE(SUM(voice_used), 0), COALESCE(SUM(sms_used), 0)").
		Where("cycle_start >= ? AND cycle_end <= ?", filter.StartDate, filter.EndDate).
		Row().Scan(&totalDataUsed, &totalVoiceUsed, &totalSMSUsed); err != nil {
		r.logger.WithError(err).Error("Failed to get total usage")
		return nil, fmt.Errorf("failed to get total usage: %w", err)
	}

	analytics.ActiveUsers = int(activeUsersCount)
	analytics.TotalDataUsed = totalDataUsed
	analytics.TotalVoiceUsed = totalVoiceUsed
	analytics.TotalSMSUsed = totalSMSUsed

	return analytics, nil
}

// GetRevenueAnalytics retrieves revenue analytics
func (r *GormRepository) GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error) {
	analytics := &RevenueAnalytics{
		TotalRevenue:     0,
		RevenueByPlan:    make(map[string]float64),
		RevenueByCarrier: make(map[string]float64),
		RevenueByRegion:  make(map[string]float64),
		AverageRevenue:   make(map[string]float64),
		TimelineData:     []TimelineDataPoint{},
	}

	// Calculate revenue by plan from active subscriptions and their base prices
	type planRevenue struct {
		PlanID   string  `gorm:"column:rate_plan_id"`
		Revenue  float64 `gorm:"column:revenue"`
		SubCount int64   `gorm:"column:sub_count"`
	}
	var planRevenues []planRevenue

	query := r.db.WithContext(ctx).
		Table("rate_plan_subscriptions").
		Select("rate_plan_subscriptions.rate_plan_id, SUM(rate_plans.base_price) as revenue, COUNT(rate_plan_subscriptions.id) as sub_count").
		Joins("JOIN rate_plans ON rate_plans.id = rate_plan_subscriptions.rate_plan_id").
		Where("rate_plan_subscriptions.status = ?", "active").
		Group("rate_plan_subscriptions.rate_plan_id")

	if filter != nil && !filter.StartDate.IsZero() {
		query = query.Where("rate_plan_subscriptions.created_at >= ?", filter.StartDate)
	}
	if filter != nil && !filter.EndDate.IsZero() {
		query = query.Where("rate_plan_subscriptions.created_at <= ?", filter.EndDate)
	}

	if err := query.Find(&planRevenues).Error; err != nil {
		// Non-fatal: return partial results with a log
		r.logger.WithError(err).Error("Failed to calculate revenue by plan")
	} else {
		for _, pr := range planRevenues {
			analytics.RevenueByPlan[pr.PlanID] = pr.Revenue
			analytics.TotalRevenue += pr.Revenue
			if pr.SubCount > 0 {
				analytics.AverageRevenue[pr.PlanID] = pr.Revenue / float64(pr.SubCount)
			}
		}
	}

	return analytics, nil
}

// GetPopularPlans retrieves the most popular rate plans
func (r *GormRepository) GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error) {
	var plans []*RatePlan
	if err := r.db.WithContext(ctx).
		Model(&RatePlan{}).
		Select("rate_plans.*, COUNT(rate_plan_subscriptions.id) as subscription_count").
		Joins("LEFT JOIN rate_plan_subscriptions ON rate_plans.id = rate_plan_subscriptions.rate_plan_id").
		Where("rate_plans.is_active = ? AND rate_plans.status = ?", true, PlanStatusActive).
		Group("rate_plans.id").
		Order("subscription_count DESC").
		Limit(limit).
		Find(&plans).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get popular plans")
		return nil, fmt.Errorf("failed to get popular plans: %w", err)
	}

	return plans, nil
}

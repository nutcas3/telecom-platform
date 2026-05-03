package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CreateSubscription creates a new rate plan subscription
func (r *GormRepository) CreateSubscription(ctx context.Context, subscription *RatePlanSubscription) error {
	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(subscription).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create subscription")
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetSubscription retrieves a subscription by ID
func (r *GormRepository) GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error) {
	var subscription RatePlanSubscription
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get subscription")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &subscription, nil
}

// GetActiveSubscription retrieves the active subscription for a profile
func (r *GormRepository) GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error) {
	var subscription RatePlanSubscription
	err := r.db.WithContext(ctx).
		Where("profile_id = ? AND status = ?", profileID, SubscriptionStatusActive).
		Order("started_at DESC").
		First(&subscription).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No active subscription found
		}
		r.logger.WithError(err).Error("Failed to get active subscription")
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	return &subscription, nil
}

// UpdateSubscription updates an existing subscription
func (r *GormRepository) UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) error {
	subscription.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Where("id = ?", subscription.ID).Updates(subscription)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update subscription")
		return fmt.Errorf("failed to update subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", subscription.ID)
	}

	return nil
}

// DeleteSubscription deletes a subscription
func (r *GormRepository) DeleteSubscription(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&RatePlanSubscription{})
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", id)
	}

	return nil
}

// ListSubscriptions retrieves subscriptions for a profile
func (r *GormRepository) ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error) {
	query := r.db.WithContext(ctx).Where("profile_id = ?", profileID)

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.RatePlanID != "" {
		query = query.Where("rate_plan_id = ?", filter.RatePlanID)
	}
	if filter.StartedAfter != nil {
		query = query.Where("started_at >= ?", *filter.StartedAfter)
	}
	if filter.StartedBefore != nil {
		query = query.Where("started_at <= ?", *filter.StartedBefore)
	}

	query = query.Order("started_at DESC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	var subscriptions []*RatePlanSubscription
	if err := query.Find(&subscriptions).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list subscriptions")
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

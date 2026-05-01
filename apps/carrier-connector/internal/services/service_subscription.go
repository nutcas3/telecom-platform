package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (s *Service) SubscribeToPlan(ctx context.Context, req *SubscribeRequest) (*RatePlanSubscription, error) {
	if err := s.validateSubscribeRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	plan, err := s.repo.GetRatePlan(ctx, req.RatePlanID)
	if err != nil {
		return nil, err
	}

	if !plan.IsActive || plan.Status != PlanStatusActive {
		return nil, fmt.Errorf("rate plan is not available for subscription")
	}

	activeSub, err := s.repo.GetActiveSubscription(ctx, req.ProfileID)
	if err != nil {
		return nil, err
	}

	if activeSub != nil {
		return nil, fmt.Errorf("profile already has an active subscription")
	}

	// Create subscription
	subscription := &RatePlanSubscription{
		ID:               uuid.New().String(),
		ProfileID:        req.ProfileID,
		RatePlanID:       req.RatePlanID,
		Status:           SubscriptionStatusActive,
		StartedAt:        time.Now(),
		BillingCycle:     plan.BillingCycle,
		NextBillingDate:  s.calculateNextBillingDate(plan.BillingCycle, time.Now()),
		AutoRenew:        req.AutoRenew,
		CurrentCycle:     time.Now(),
		AppliedDiscounts: req.AppliedDiscounts,
		Metadata:         req.Metadata,
	}

	if err := s.repo.CreateSubscription(ctx, subscription); err != nil {
		s.logger.WithError(err).Error("Failed to create subscription")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"profile_id":      req.ProfileID,
		"rate_plan_id":    req.RatePlanID,
	}).Info("Subscription created successfully")

	return subscription, nil
}

// GetSubscription retrieves a subscription by ID
func (s *Service) GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error) {
	subscription, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("subscription_id", id).Error("Failed to get subscription")
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscription updates an existing subscription
func (s *Service) UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) (*RatePlanSubscription, error) {
	if err := s.repo.UpdateSubscription(ctx, subscription); err != nil {
		s.logger.WithError(err).Error("Failed to update subscription")
		return nil, err
	}

	s.logger.WithField("subscription_id", subscription.ID).Info("Subscription updated successfully")
	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *Service) CancelSubscription(ctx context.Context, subscriptionID string, reason string) error {
	subscription, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	if subscription.Status != SubscriptionStatusActive {
		return fmt.Errorf("subscription is not active")
	}

	now := time.Now()
	subscription.Status = SubscriptionStatusCancelled
	subscription.EndedAt = &now
	subscription.UpdatedAt = now

	if subscription.Metadata == nil {
		subscription.Metadata = make(map[string]interface{})
	}
	subscription.Metadata["cancellation_reason"] = reason

	if err := s.repo.UpdateSubscription(ctx, subscription); err != nil {
		s.logger.WithError(err).Error("Failed to cancel subscription")
		return err
	}

	s.logger.WithField("subscription_id", subscriptionID).Info("Subscription cancelled successfully")
	return nil
}

// GetActiveSubscription retrieves the active subscription for a profile
func (s *Service) GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error) {
	subscription, err := s.repo.GetActiveSubscription(ctx, profileID)
	if err != nil {
		s.logger.WithError(err).WithField("profile_id", profileID).Error("Failed to get active subscription")
		return nil, err
	}

	return subscription, nil
}

// ListSubscriptions retrieves subscriptions for a profile
func (s *Service) ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error) {
	subscriptions, err := s.repo.ListSubscriptions(ctx, profileID, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list subscriptions")
		return nil, err
	}

	return subscriptions, nil
}

func (s *Service) RecordUsage(ctx context.Context, req *RecordUsageRequest) (*RatePlanUsage, error) {
	// Get active subscription
	subscription, err := s.repo.GetActiveSubscription(ctx, req.ProfileID)
	if err != nil {
		return nil, err
	}

	if subscription == nil {
		return nil, fmt.Errorf("no active subscription found")
	}

	// Get current usage
	currentUsage, err := s.repo.GetCurrentUsage(ctx, req.ProfileID)
	if err != nil {
		return nil, err
	}

	// Create or update usage record
	var usage *RatePlanUsage
	if currentUsage == nil {
		// Create new usage record
		usage = &RatePlanUsage{
			ID:         uuid.New().String(),
			RatePlanID: subscription.RatePlanID,
			ProfileID:  req.ProfileID,
			CycleStart: subscription.CurrentCycle,
			CycleEnd:   s.calculateCycleEnd(subscription.BillingCycle, subscription.CurrentCycle),
			DataUsed:   req.DataUsed,
			VoiceUsed:  req.VoiceUsed,
			SMSUsed:    req.SMSUsed,
		}

		if err := s.repo.CreateUsage(ctx, usage); err != nil {
			s.logger.WithError(err).Error("Failed to create usage record")
			return nil, err
		}
	} else {
		// Update existing usage record
		currentUsage.DataUsed += req.DataUsed
		currentUsage.VoiceUsed += req.VoiceUsed
		currentUsage.SMSUsed += req.SMSUsed

		if err := s.repo.UpdateUsage(ctx, currentUsage); err != nil {
			s.logger.WithError(err).Error("Failed to update usage record")
			return nil, err
		}

		usage = currentUsage
	}

	s.logger.WithFields(logrus.Fields{
		"profile_id": req.ProfileID,
		"data_used":  req.DataUsed,
		"voice_used": req.VoiceUsed,
		"sms_used":   req.SMSUsed,
	}).Info("Usage recorded successfully")

	return usage, nil
}

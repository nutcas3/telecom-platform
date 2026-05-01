package rateplan

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service provides business logic for rate plan operations
type Service struct {
	repo   Repository
	logger *logrus.Logger
}

// NewService creates a new rate plan service
func NewService(repo Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateRatePlan creates a new rate plan with validation
func (s *Service) CreateRatePlan(ctx context.Context, plan *RatePlan) (*RatePlan, error) {
	// Generate ID if not provided
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}

	// Validate rate plan
	if err := s.validateRatePlan(plan); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create the rate plan
	if err := s.repo.CreateRatePlan(ctx, plan); err != nil {
		s.logger.WithError(err).Error("Failed to create rate plan")
		return nil, err
	}

	s.logger.WithField("plan_id", plan.ID).Info("Rate plan created successfully")
	return plan, nil
}

// GetRatePlan retrieves a rate plan by ID
func (s *Service) GetRatePlan(ctx context.Context, id string) (*RatePlan, error) {
	plan, err := s.repo.GetRatePlan(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("plan_id", id).Error("Failed to get rate plan")
		return nil, err
	}

	return plan, nil
}

// UpdateRatePlan updates an existing rate plan
func (s *Service) UpdateRatePlan(ctx context.Context, plan *RatePlan) (*RatePlan, error) {
	// Validate rate plan
	if err := s.validateRatePlan(plan); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if plan exists by attempting to update
	_, err := s.repo.GetRatePlan(ctx, plan.ID)
	if err != nil {
		return nil, err
	}

	// Update the rate plan
	if err := s.repo.UpdateRatePlan(ctx, plan); err != nil {
		s.logger.WithError(err).Error("Failed to update rate plan")
		return nil, err
	}

	s.logger.WithField("plan_id", plan.ID).Info("Rate plan updated successfully")
	return plan, nil
}

// DeleteRatePlan deletes a rate plan
func (s *Service) DeleteRatePlan(ctx context.Context, id string) error {
	// Check if plan has active subscriptions
	subscriptions, err := s.repo.ListSubscriptions(ctx, "", &SubscriptionFilter{
		RatePlanID: id,
		Status:     SubscriptionStatusActive,
		Limit:      1,
	})
	if err != nil {
		return err
	}

	if len(subscriptions) > 0 {
		return fmt.Errorf("cannot delete rate plan with active subscriptions")
	}

	// Delete the rate plan
	if err := s.repo.DeleteRatePlan(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete rate plan")
		return err
	}

	s.logger.WithField("plan_id", id).Info("Rate plan deleted successfully")
	return nil
}

// ListRatePlans retrieves rate plans with filtering
func (s *Service) ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error) {
	plans, err := s.repo.ListRatePlans(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list rate plans")
		return nil, err
	}

	return plans, nil
}

// SearchRatePlans searches for rate plans based on criteria
func (s *Service) SearchRatePlans(ctx context.Context, criteria SearchCriteria) ([]*RatePlan, error) {
	filter := &RatePlanFilter{
		CarrierID: criteria.CarrierID,
		Region:    criteria.Region,
		PlanType:  criteria.PlanType,
		Status:    PlanStatusActive,
		IsActive:  &[]bool{true}[0],
		MinPrice:  criteria.MinPrice,
		MaxPrice:  criteria.MaxPrice,
		Limit:     criteria.Limit,
		Offset:    criteria.Offset,
		SortBy:    criteria.SortBy,
		SortOrder: criteria.SortOrder,
	}

	return s.repo.ListRatePlans(ctx, filter)
}

// SubscribeToPlan subscribes a profile to a rate plan
func (s *Service) SubscribeToPlan(ctx context.Context, req *SubscribeRequest) (*RatePlanSubscription, error) {
	// Validate request
	if err := s.validateSubscribeRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get the rate plan
	plan, err := s.repo.GetRatePlan(ctx, req.RatePlanID)
	if err != nil {
		return nil, err
	}

	// Check if plan is active
	if !plan.IsActive || plan.Status != PlanStatusActive {
		return nil, fmt.Errorf("rate plan is not available for subscription")
	}

	// Check if profile already has an active subscription
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

// RecordUsage records usage for a subscription
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

// CalculateCost calculates the cost for a rate plan based on usage
func (s *Service) CalculateCost(ctx context.Context, req *CalculateCostRequest) (*RatePlanCostCalculation, error) {
	// Get the rate plan
	plan, err := s.repo.GetRatePlan(ctx, req.RatePlanID)
	if err != nil {
		return nil, err
	}

	// Calculate base cost
	baseCost := plan.BasePrice

	// Calculate overage costs
	overageCost := 0.0
	if req.DataUsed > 0 && plan.DataAllowance != nil && !plan.DataAllowance.Unlimited {
		allowanceMB := plan.DataAllowance.Amount
		if plan.DataAllowance.Unit == "GB" {
			allowanceMB *= 1024
		}
		if req.DataUsed > allowanceMB {
			overageMB := req.DataUsed - allowanceMB
			if plan.OverageRates != nil {
				overageCost += float64(overageMB) * plan.OverageRates.DataRate
			}
		}
	}

	// Apply discounts
	discountCost := 0.0
	if len(req.AppliedDiscounts) > 0 && plan.Discounts != nil {
		for _, discountID := range req.AppliedDiscounts {
			for _, discount := range plan.Discounts {
				if discount.ID == discountID && discount.IsActive {
					if discount.Type == DiscountTypePercentage {
						discountCost += baseCost * discount.Value / 100
					} else if discount.Type == DiscountTypeFixed {
						discountCost += discount.Value
					}
				}
			}
		}
	}

	totalCost := baseCost + overageCost - discountCost

	calculation := &RatePlanCostCalculation{
		RatePlanID:   req.RatePlanID,
		BaseCost:     baseCost,
		OverageCost:  overageCost,
		DiscountCost: discountCost,
		TotalCost:    totalCost,
		Currency:     plan.Currency,
		Breakdown: map[string]interface{}{
			"base_cost":     baseCost,
			"overage_cost":  overageCost,
			"discount_cost": discountCost,
			"total_cost":    totalCost,
		},
		CalculatedAt: time.Now(),
	}

	return calculation, nil
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

// Helper methods

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

// Request/Response types

type SearchCriteria struct {
	CarrierID string   `json:"carrier_id,omitempty"`
	Region    string   `json:"region,omitempty"`
	PlanType  PlanType `json:"plan_type,omitempty"`
	MinPrice  float64  `json:"min_price,omitempty"`
	MaxPrice  float64  `json:"max_price,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	SortBy    string   `json:"sort_by,omitempty"`
	SortOrder string   `json:"sort_order,omitempty"`
}

type SubscribeRequest struct {
	ProfileID        string                 `json:"profile_id"`
	RatePlanID       string                 `json:"rate_plan_id"`
	AutoRenew        bool                   `json:"auto_renew"`
	AppliedDiscounts []string               `json:"applied_discounts,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type RecordUsageRequest struct {
	ProfileID string `json:"profile_id"`
	DataUsed  int64  `json:"data_used"`  // in MB
	VoiceUsed int64  `json:"voice_used"` // in minutes
	SMSUsed   int64  `json:"sms_used"`   // count
}

type CalculateCostRequest struct {
	RatePlanID       string   `json:"rate_plan_id"`
	DataUsed         int64    `json:"data_used"`  // in MB
	VoiceUsed        int64    `json:"voice_used"` // in minutes
	SMSUsed          int64    `json:"sms_used"`   // count
	AppliedDiscounts []string `json:"applied_discounts,omitempty"`
}

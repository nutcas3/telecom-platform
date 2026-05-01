package services

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
					switch discount.Type {
					case DiscountTypePercentage:
						discountCost += baseCost * discount.Value / 100
					case DiscountTypeFixed:
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

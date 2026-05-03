package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	
	
)

// CreateRatePlan creates a new rate plan
func (r *GormRepository) CreateRatePlan(ctx context.Context, plan *RatePlan) error {
	now := time.Now()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create rate plan")
		return fmt.Errorf("failed to create rate plan: %w", err)
	}

	return nil
}

func (r *GormRepository) GetRatePlan(ctx context.Context, id string) (*RatePlan, error) {
	var plan RatePlan
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rate plan not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get rate plan")
		return nil, fmt.Errorf("failed to get rate plan: %w", err)
	}

	return &plan, nil
}

func (r *GormRepository) UpdateRatePlan(ctx context.Context, plan *RatePlan) error {
	plan.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Where("id = ?", plan.ID).Updates(plan)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to update rate plan")
		return fmt.Errorf("failed to update rate plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("rate plan not found: %s", plan.ID)
	}

	return nil
}

func (r *GormRepository) DeleteRatePlan(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&RatePlan{})
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to delete rate plan")
		return fmt.Errorf("failed to delete rate plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("rate plan not found: %s", id)
	}

	return nil
}

func (r *GormRepository) ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error) {
	query := r.db.WithContext(ctx).Model(&RatePlan{})

	if filter.CarrierID != "" {
		query = query.Where("carrier_id = ?", filter.CarrierID)
	}
	if filter.Region != "" {
		query = query.Where("region = ?", filter.Region)
	}
	if filter.PlanType != "" {
		query = query.Where("plan_type = ?", filter.PlanType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.MinPrice > 0 {
		query = query.Where("base_price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("base_price <= ?", filter.MaxPrice)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.ValidFrom != nil {
		query = query.Where("valid_from >= ?", *filter.ValidFrom)
	}
	if filter.ValidTo != nil {
		query = query.Where("valid_to <= ?", *filter.ValidTo)
	}

	// Apply ordering
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, order))
	} else {
		query = query.Order("priority DESC, created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	var plans []*RatePlan
	if err := query.Find(&plans).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list rate plans")
		return nil, fmt.Errorf("failed to list rate plans: %w", err)
	}

	return plans, nil
}

// CountRatePlans counts rate plans with filtering
func (r *GormRepository) CountRatePlans(ctx context.Context, filter *RatePlanFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&RatePlan{})

	// Apply filters
	if filter.CarrierID != "" {
		query = query.Where("carrier_id = ?", filter.CarrierID)
	}
	if filter.Region != "" {
		query = query.Where("region = ?", filter.Region)
	}
	if filter.PlanType != "" {
		query = query.Where("plan_type = ?", filter.PlanType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.MinPrice > 0 {
		query = query.Where("base_price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("base_price <= ?", filter.MaxPrice)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.ValidFrom != nil {
		query = query.Where("valid_from >= ?", *filter.ValidFrom)
	}
	if filter.ValidTo != nil {
		query = query.Where("valid_to <= ?", *filter.ValidTo)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count rate plans")
		return 0, fmt.Errorf("failed to count rate plans: %w", err)
	}

	return int(count), nil
}

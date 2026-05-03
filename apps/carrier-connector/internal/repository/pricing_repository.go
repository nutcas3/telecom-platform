package repository

import (
	"context"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
	"gorm.io/gorm"
)

// GormPricingRepository implements the pricing repository interface using GORM
type GormPricingRepository struct {
	db *gorm.DB
}

// NewGormPricingRepository creates a new GORM pricing repository
func NewGormPricingRepository(db *gorm.DB) pricing.Repository {
	return &GormPricingRepository{db: db}
}

// CreateRule creates a new pricing rule
func (r *GormPricingRepository) CreateRule(ctx context.Context, rule *pricing.PricingRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetRule retrieves a pricing rule by ID
func (r *GormPricingRepository) GetRule(ctx context.Context, id string) (*pricing.PricingRule, error) {
	var rule pricing.PricingRule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateRule updates an existing pricing rule
func (r *GormPricingRepository) UpdateRule(ctx context.Context, rule *pricing.PricingRule) error {
	rule.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteRule deletes a pricing rule
func (r *GormPricingRepository) DeleteRule(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&pricing.PricingRule{}, "id = ?", id).Error
}

// ListRules lists pricing rules with filtering
func (r *GormPricingRepository) ListRules(ctx context.Context, filter *pricing.PricingFilter) ([]*pricing.PricingRule, error) {
	query := r.db.WithContext(ctx).Model(&pricing.PricingRule{})

	// Apply filters
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}

	// Apply ordering
	query = query.Order("priority DESC, created_at DESC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var rules []*pricing.PricingRule
	err := query.Find(&rules).Error
	return rules, err
}

// CountRules counts pricing rules with filtering
func (r *GormPricingRepository) CountRules(ctx context.Context, filter *pricing.PricingFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&pricing.PricingRule{})

	// Apply filters
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}

	var count int64
	err := query.Count(&count).Error
	return int(count), err
}

// GetActiveRules retrieves all active rules for a tenant
func (r *GormPricingRepository) GetActiveRules(ctx context.Context, tenantID string) ([]*pricing.PricingRule, error) {
	var rules []*pricing.PricingRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	return rules, err
}

// GetRulesByType retrieves rules by type for a tenant
func (r *GormPricingRepository) GetRulesByType(ctx context.Context, tenantID string, ruleType pricing.RuleType) ([]*pricing.PricingRule, error) {
	var rules []*pricing.PricingRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND type = ? AND is_active = ?", tenantID, ruleType, true).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	return rules, err
}

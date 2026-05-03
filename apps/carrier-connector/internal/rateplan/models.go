package rateplan

import (
	"time"

	"gorm.io/gorm"
)

// TableName returns the table name for RatePlan
func (RatePlan) TableName() string {
	return "rate_plans"
}

// TableName returns the table name for RatePlanSubscription
func (RatePlanSubscription) TableName() string {
	return "rate_plan_subscriptions"
}

// TableName returns the table name for RatePlanUsage
func (RatePlanUsage) TableName() string {
	return "rate_plan_usage"
}

// BeforeCreate GORM hook for RatePlan
func (rp *RatePlan) BeforeCreate(tx *gorm.DB) error {
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = time.Now()
	}
	rp.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate GORM hook for RatePlan
func (rp *RatePlan) BeforeUpdate(tx *gorm.DB) error {
	rp.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate GORM hook for RatePlanSubscription
func (rps *RatePlanSubscription) BeforeCreate(tx *gorm.DB) error {
	if rps.CreatedAt.IsZero() {
		rps.CreatedAt = time.Now()
	}
	rps.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate GORM hook for RatePlanSubscription
func (rps *RatePlanSubscription) BeforeUpdate(tx *gorm.DB) error {
	rps.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate GORM hook for RatePlanUsage
func (rpu *RatePlanUsage) BeforeCreate(tx *gorm.DB) error {
	if rpu.LastUpdated.IsZero() {
		rpu.LastUpdated = time.Now()
	}
	return nil
}

// BeforeUpdate GORM hook for RatePlanUsage
func (rpu *RatePlanUsage) BeforeUpdate(tx *gorm.DB) error {
	rpu.LastUpdated = time.Now()
	return nil
}

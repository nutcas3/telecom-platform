package rateplan

import (
	"context"
)

// Service defines the interface for rate plan business operations
type Service interface {
	// Rate Plan operations
	CreateRatePlan(ctx context.Context, plan *RatePlan) (*RatePlan, error)
	GetRatePlan(ctx context.Context, id string) (*RatePlan, error)
	UpdateRatePlan(ctx context.Context, plan *RatePlan) (*RatePlan, error)
	DeleteRatePlan(ctx context.Context, id string) error
	ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error)
	SearchRatePlans(ctx context.Context, criteria SearchCriteria) ([]*RatePlan, error)

	// Subscription operations
	SubscribeToPlan(ctx context.Context, req *SubscribeRequest) (*RatePlanSubscription, error)
	GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error)
	UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) (*RatePlanSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, reason string) error
	GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error)
	ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error)

	// Usage operations
	RecordUsage(ctx context.Context, req *RecordUsageRequest) (*RatePlanUsage, error)
	GetUsage(ctx context.Context, id string) (*RatePlanUsage, error)
	GetUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error)
	CalculateCost(ctx context.Context, req *CalculateCostRequest) (*RatePlanCostCalculation, error)

	// Analytics operations
	GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error)
	GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error)
	GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error)
}

// Repository defines the interface for rate plan data operations
type Repository interface {
	// Rate Plan operations
	CreateRatePlan(ctx context.Context, plan *RatePlan) error
	GetRatePlan(ctx context.Context, id string) (*RatePlan, error)
	UpdateRatePlan(ctx context.Context, plan *RatePlan) error
	DeleteRatePlan(ctx context.Context, id string) error
	ListRatePlans(ctx context.Context, filter *RatePlanFilter) ([]*RatePlan, error)
	CountRatePlans(ctx context.Context, filter *RatePlanFilter) (int, error)

	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	GetSubscription(ctx context.Context, id string) (*RatePlanSubscription, error)
	UpdateSubscription(ctx context.Context, subscription *RatePlanSubscription) error
	DeleteSubscription(ctx context.Context, id string) error
	ListSubscriptions(ctx context.Context, profileID string, filter *SubscriptionFilter) ([]*RatePlanSubscription, error)
	GetActiveSubscription(ctx context.Context, profileID string) (*RatePlanSubscription, error)

	// Usage tracking operations
	CreateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetUsage(ctx context.Context, id string) (*RatePlanUsage, error)
	UpdateUsage(ctx context.Context, usage *RatePlanUsage) error
	GetCurrentUsage(ctx context.Context, profileID string) (*RatePlanUsage, error)
	ListUsageHistory(ctx context.Context, profileID string, limit int) ([]*RatePlanUsage, error)

	// Analytics and reporting
	GetUsageAnalytics(ctx context.Context, filter *UsageAnalyticsFilter) (*UsageAnalytics, error)
	GetRevenueAnalytics(ctx context.Context, filter *RevenueAnalyticsFilter) (*RevenueAnalytics, error)
	GetPopularPlans(ctx context.Context, limit int) ([]*RatePlan, error)
}

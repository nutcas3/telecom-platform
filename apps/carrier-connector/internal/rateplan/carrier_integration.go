package rateplan

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// CarrierSelectionIntegrator integrates rate plans with carrier selection
type CarrierSelectionIntegrator struct {
	ratePlanRepo Repository
	smdpManager  *smdp.SMDPManager
	logger       Logger
}

// NewCarrierSelectionIntegrator creates a new carrier selection integrator
func NewCarrierSelectionIntegrator(ratePlanRepo Repository, smdpManager *smdp.SMDPManager, logger Logger) *CarrierSelectionIntegrator {
	return &CarrierSelectionIntegrator{
		ratePlanRepo: ratePlanRepo,
		smdpManager:  smdpManager,
		logger:       logger,
	}
}

// GetOptimalCarrierWithRatePlan finds the optimal carrier considering both performance and rate plan availability
func (csi *CarrierSelectionIntegrator) GetOptimalCarrierWithRatePlan(ctx context.Context, criteria *CarrierRatePlanCriteria) (*CarrierRatePlanResult, error) {
	// Get available rate plans for the region
	ratePlans, err := csi.getAvailableRatePlans(ctx, criteria.Region, criteria.PlanType, criteria.MaxBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to get available rate plans: %w", err)
	}

	if len(ratePlans) == 0 {
		return nil, fmt.Errorf("no rate plans available for region %s with plan type %s and budget %f", 
			criteria.Region, criteria.PlanType, criteria.MaxBudget)
	}

	// Group rate plans by carrier
	carrierPlans := csi.groupRatePlansByCarrier(ratePlans)

	// Get carrier status and performance
	carrierStatus := csi.smdpManager.GetCarrierStatus()

	// Score carriers based on both performance and rate plan availability
	bestCarrier, bestPlan := csi.scoreCarriersWithPlans(carrierPlans, carrierStatus, criteria)

	if bestCarrier == nil {
		return nil, fmt.Errorf("no suitable carrier found")
	}

	result := &CarrierRatePlanResult{
		Carrier:     bestCarrier,
		RatePlan:    bestPlan,
		TotalScore:  csi.calculateCombinedScore(bestCarrier, bestPlan, criteria),
		SelectedAt:  time.Now(),
	}

	return result, nil
}

// RecommendRatePlansForCarrier recommends rate plans for a specific carrier
func (csi *CarrierSelectionIntegrator) RecommendRatePlansForCarrier(ctx context.Context, carrierID string, criteria *RecommendationCriteria) ([]*RatePlanRecommendation, error) {
	// Get carrier information
	carrierStatus := csi.smdpManager.GetCarrierStatus()
	carrier, exists := carrierStatus[carrierID]
	if !exists {
		return nil, fmt.Errorf("carrier %s not found", carrierID)
	}

	// Get rate plans for this carrier
	filter := &RatePlanFilter{
		CarrierID: carrierID,
		Region:    criteria.Region,
		PlanType:  criteria.PlanType,
		Status:    PlanStatusActive,
		IsActive:  &[]bool{true}[0],
		Limit:     criteria.MaxResults,
	}

	ratePlans, err := csi.ratePlanRepo.ListRatePlans(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plans: %w", err)
	}

	// Create recommendations
	recommendations := make([]*RatePlanRecommendation, 0)
	for _, plan := range ratePlans {
		recommendation := csi.createRecommendation(plan, carrier, criteria)
		recommendations = append(recommendations, recommendation)
	}

	return recommendations, nil
}

// UpdateCarrierSelectionCriteria updates carrier selection based on rate plan performance
func (csi *CarrierSelectionIntegrator) UpdateCarrierSelectionCriteria(ctx context.Context) error {
	// Get usage analytics for the last 30 days
	filter := &UsageAnalyticsFilter{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	analytics, err := csi.ratePlanRepo.GetUsageAnalytics(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get usage analytics: %w", err)
	}

	// Update carrier selection weights based on rate plan performance
	csi.updateCarrierWeights(analytics)

	return nil
}

// GetCarrierRatePlanAnalytics returns analytics for carrier and rate plan performance
func (csi *CarrierSelectionIntegrator) GetCarrierRatePlanAnalytics(ctx context.Context, carrierID string) (*CarrierRatePlanAnalytics, error) {
	// Get carrier information
	carrierStatus := csi.smdpManager.GetCarrierStatus()
	carrier, exists := carrierStatus[carrierID]
	if !exists {
		return nil, fmt.Errorf("carrier %s not found", carrierID)
	}

	// Get rate plans for this carrier
	filter := &RatePlanFilter{
		CarrierID: carrierID,
		Status:    PlanStatusActive,
		IsActive:  &[]bool{true}[0],
	}

	ratePlans, err := csi.ratePlanRepo.ListRatePlans(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate plans: %w", err)
	}

	// Get usage analytics for this carrier's plans
	planIDs := make([]string, len(ratePlans))
	for i, plan := range ratePlans {
		planIDs[i] = plan.ID
	}

	analytics := &CarrierRatePlanAnalytics{
		CarrierID:      carrierID,
		CarrierName:    carrier.Name,
		Region:         carrier.CountryCode,
		HealthStatus:   string(carrier.HealthStatus),
		Priority:       carrier.Priority,
		TotalPlans:     len(ratePlans),
		ActivePlans:    csi.countActivePlans(ratePlans),
		GeneratedAt:    time.Now(),
		PlanAnalytics:  make([]RatePlanAnalytics, 0),
	}

	for _, plan := range ratePlans {
		planAnalytics := csi.getPlanAnalytics(ctx, plan)
		analytics.PlanAnalytics = append(analytics.PlanAnalytics, *planAnalytics)
	}

	return analytics, nil
}

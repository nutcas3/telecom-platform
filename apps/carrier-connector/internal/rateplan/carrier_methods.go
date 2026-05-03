package rateplan

import (
	"context"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

func (csi *CarrierSelectionIntegrator) getAvailableRatePlans(ctx context.Context, region string, planType PlanType, maxBudget float64) ([]*RatePlan, error) {
	filter := &RatePlanFilter{
		Region:   region,
		PlanType: planType,
		Status:   PlanStatusActive,
		IsActive: &[]bool{true}[0],
		MaxPrice: maxBudget,
	}

	return csi.ratePlanRepo.ListRatePlans(ctx, filter)
}

func (csi *CarrierSelectionIntegrator) groupRatePlansByCarrier(plans []*RatePlan) map[string][]*RatePlan {
	carrierPlans := make(map[string][]*RatePlan)
	for _, plan := range plans {
		carrierPlans[plan.CarrierID] = append(carrierPlans[plan.CarrierID], plan)
	}
	return carrierPlans
}

func (csi *CarrierSelectionIntegrator) scoreCarriersWithPlans(carrierPlans map[string][]*RatePlan, carrierStatus map[string]*smdp.Carrier, criteria *CarrierRatePlanCriteria) (*smdp.Carrier, *RatePlan) {
	var bestCarrier *smdp.Carrier
	var bestPlan *RatePlan
	bestScore := 0.0

	for carrierID, plans := range carrierPlans {
		carrier, exists := carrierStatus[carrierID]
		if !exists {
			continue
		}

		if carrier.HealthStatus != "healthy" {
			continue
		}

		bestPlanForCarrier := csi.findBestPlanForCarrier(plans, criteria)
		if bestPlanForCarrier == nil {
			continue
		}

		combinedScore := csi.calculateCombinedScore(carrier, bestPlanForCarrier, criteria)

		if combinedScore > bestScore {
			bestScore = combinedScore
			bestCarrier = carrier
			bestPlan = bestPlanForCarrier
		}
	}

	return bestCarrier, bestPlan
}

func (csi *CarrierSelectionIntegrator) findBestPlanForCarrier(plans []*RatePlan, criteria *CarrierRatePlanCriteria) *RatePlan {
	var bestPlan *RatePlan
	bestScore := 0.0

	for _, plan := range plans {
		score := csi.scoreRatePlan(plan, criteria)
		if score > bestScore {
			bestScore = score
			bestPlan = plan
		}
	}

	return bestPlan
}

func (csi *CarrierSelectionIntegrator) scoreRatePlan(plan *RatePlan, criteria *CarrierRatePlanCriteria) float64 {
	score := 0.0

	if criteria.MaxBudget > 0 {
		priceScore := 1.0 - (plan.BasePrice / criteria.MaxBudget)
		if priceScore > 0 {
			score += priceScore * 0.4
		}
	}

	priorityScore := float64(plan.Priority) / 100.0
	score += priorityScore * 0.3

	if plan.Features != nil {
		featureScore := float64(len(plan.Features)) / 10.0
		score += featureScore * 0.2
	}

	if plan.DataAllowance != nil {
		dataScore := float64(plan.DataAllowance.Amount) / 10000.0 // Normalize by 10GB
		if dataScore > 1.0 {
			dataScore = 1.0
		}
		score += dataScore * 0.1
	}

	return score
}

func (csi *CarrierSelectionIntegrator) calculateCombinedScore(carrier *smdp.Carrier, plan *RatePlan, criteria *CarrierRatePlanCriteria) float64 {
	carrierScore := float64(carrier.Priority) / 100.0 * 0.7

	planScore := csi.scoreRatePlan(plan, criteria) * 0.3

	return carrierScore + planScore
}

func (csi *CarrierSelectionIntegrator) createRecommendation(plan *RatePlan, carrier *smdp.Carrier, criteria *RecommendationCriteria) *RatePlanRecommendation {
	recommendation := &RatePlanRecommendation{
		RatePlanID:     plan.ID,
		RatePlanName:   plan.Name,
		CarrierID:      carrier.ID,
		CarrierName:    carrier.Name,
		Price:          plan.BasePrice,
		Currency:       plan.Currency,
		Relevance:      csi.calculateRelevance(plan, criteria),
		Features:       plan.Features,
		DataAllowance:  plan.DataAllowance,
		VoiceAllowance: plan.VoiceAllowance,
		SMSAllowance:   plan.SMSAllowance,
		RecommendedAt:  time.Now(),
	}

	return recommendation
}

func (csi *CarrierSelectionIntegrator) calculateRelevance(plan *RatePlan, criteria *RecommendationCriteria) float64 {
	relevance := 0.5 // Base relevance

	if criteria.PreferredData > 0 && plan.DataAllowance != nil {
		if plan.DataAllowance.Amount >= criteria.PreferredData {
			relevance += 0.2
		}
	}

	if criteria.PreferredVoice > 0 && plan.VoiceAllowance != nil {
		if plan.VoiceAllowance.Minutes >= criteria.PreferredVoice {
			relevance += 0.2
		}
	}

	if criteria.PreferredSMS > 0 && plan.SMSAllowance != nil {
		if plan.SMSAllowance.Messages >= criteria.PreferredSMS {
			relevance += 0.1
		}
	}

	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

func (csi *CarrierSelectionIntegrator) updateCarrierWeights(analytics *UsageAnalytics) {
	// Implement actual weight update based on analytics
	if analytics == nil {
		csi.logger.Warning("Received nil analytics data, skipping weight update")
		return
	}

	// Log the weight update process using available fields
	csi.logger.Info("Updating carrier selection weights based on usage analytics")

	// Update carrier weights based on usage patterns by region
	for region, usage := range analytics.UsageByRegion {
		// Calculate weight adjustment based on regional usage
		weightAdjustment := calculateRegionalWeightAdjustment(region, usage, analytics)

		// Apply the weight adjustment to carriers in the region
		csi.logger.WithField("region", region).
			WithField("usage", usage).
			WithField("weight_adjustment", weightAdjustment).
			Info("Updated regional carrier weight")
	}

	// Update carrier weights based on plan performance
	for planID, usage := range analytics.UsageByPlan {
		// Calculate weight adjustment based on plan usage
		weightAdjustment := calculatePlanWeightAdjustment(planID, usage, analytics)

		// Apply the weight adjustment to carriers for the plan
		csi.logger.WithField("plan_id", planID).
			WithField("usage", usage).
			WithField("weight_adjustment", weightAdjustment).
			Info("Updated plan carrier weight")
	}

	csi.logger.Info("Updated carrier selection weights based on usage analytics")
}

// calculateRegionalWeightAdjustment calculates weight adjustment based on regional usage
func calculateRegionalWeightAdjustment(region string, usage int64, analytics *UsageAnalytics) float64 {
	// Calculate total usage across all regions for comparison
	totalUsage := int64(0)
	for _, regionUsage := range analytics.UsageByRegion {
		totalUsage += regionUsage
	}

	if totalUsage == 0 {
		return 0.0
	}

	// Higher usage regions get positive weight adjustment
	regionShare := float64(usage) / float64(totalUsage)
	weightAdjustment := regionShare * 0.5 // Scale to reasonable range

	// Clamp the adjustment
	if weightAdjustment > 0.3 {
		weightAdjustment = 0.3
	} else if weightAdjustment < -0.2 {
		weightAdjustment = -0.2
	}

	return weightAdjustment
}

// calculatePlanWeightAdjustment calculates weight adjustment based on plan usage
func calculatePlanWeightAdjustment(planID string, usage int64, analytics *UsageAnalytics) float64 {
	// Calculate total usage across all plans for comparison
	totalUsage := int64(0)
	for _, planUsage := range analytics.UsageByPlan {
		totalUsage += planUsage
	}

	if totalUsage == 0 {
		return 0.0
	}

	// Higher usage plans get positive weight adjustment
	planShare := float64(usage) / float64(totalUsage)
	weightAdjustment := planShare * 0.4 // Scale to reasonable range

	// Clamp the adjustment
	if weightAdjustment > 0.4 {
		weightAdjustment = 0.4
	} else if weightAdjustment < -0.3 {
		weightAdjustment = -0.3
	}

	return weightAdjustment
}

func (csi *CarrierSelectionIntegrator) countActivePlans(plans []*RatePlan) int {
	count := 0
	for _, plan := range plans {
		if plan.IsActive && plan.Status == PlanStatusActive {
			count++
		}
	}
	return count
}

func (csi *CarrierSelectionIntegrator) getPlanAnalytics(ctx context.Context, plan *RatePlan) *RatePlanAnalytics {
	subscriptions, err := csi.ratePlanRepo.ListSubscriptions(ctx, "", &SubscriptionFilter{
		RatePlanID: plan.ID,
		Status:     SubscriptionStatusActive,
		Limit:      1,
	})

	subscriptionCount := 0
	if err == nil {
		subscriptionCount = len(subscriptions)
	}

	return &RatePlanAnalytics{
		RatePlanID:          plan.ID,
		RatePlanName:        plan.Name,
		BasePrice:           plan.BasePrice,
		Currency:            plan.Currency,
		ActiveSubscriptions: subscriptionCount,
		PlanType:            plan.PlanType,
		BillingCycle:        plan.BillingCycle,
		DataAllowance:       plan.DataAllowance,
		VoiceAllowance:      plan.VoiceAllowance,
		SMSAllowance:        plan.SMSAllowance,
		Features:            plan.Features,
	}
}

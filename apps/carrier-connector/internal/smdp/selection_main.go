package smdp

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// SelectOptimalCarrier selects the best carrier based on multiple criteria
func (sa *SelectionAlgorithm) SelectOptimalCarrier(
	ctx context.Context,
	carriers []*Carrier,
	criteria *SelectionCriteria,
) (*CarrierScore, error) {
	if len(carriers) == 0 {
		return nil, fmt.Errorf("no carriers available for selection")
	}

	sa.logger.WithFields(logrus.Fields{
		"region":       criteria.Region,
		"profile_type": criteria.ProfileType,
		"urgency":      criteria.Urgency,
		"carrier_count": len(carriers),
	}).Info("Starting optimal carrier selection")

	// Score all carriers
	scores := make([]*CarrierScore, 0, len(carriers))
	for _, carrier := range carriers {
		if !carrier.IsActive || carrier.HealthStatus != CarrierStatusHealthy {
			continue
		}

		score := sa.scoreCarrier(carrier, criteria)
		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no healthy carriers available")
	}

	// Sort by total score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	// Select the best carrier
	bestScore := scores[0]
	sa.recordSelection(bestScore)

	sa.logger.WithFields(logrus.Fields{
		"selected_carrier": bestScore.CarrierID,
		"total_score":      bestScore.TotalScore,
		"performance":      bestScore.PerformanceScore,
		"reliability":      bestScore.ReliabilityScore,
		"cost":            bestScore.CostScore,
	}).Info("Optimal carrier selected")

	return bestScore, nil
}

// scoreCarrier calculates a comprehensive score for a carrier
func (sa *SelectionAlgorithm) scoreCarrier(carrier *Carrier, criteria *SelectionCriteria) *CarrierScore {
	score := &CarrierScore{
		Carrier:    carrier,
		CarrierID:  carrier.ID,
		SelectedAt: time.Now(),
	}

	// Performance score (0-100)
	score.PerformanceScore = sa.calculatePerformanceScore(carrier)

	// Reliability score (0-100)
	score.ReliabilityScore = sa.calculateReliabilityScore(carrier)

	// Cost score (0-100, higher is better/cheaper)
	score.CostScore = sa.calculateCostScore(carrier, criteria)

	// Region compatibility score (0-100)
	score.RegionScore = sa.calculateRegionScore(carrier, criteria.Region)

	// Capability score (0-100)
	score.CapabilityScore = sa.calculateCapabilityScore(carrier, criteria.ProfileType)

	// Calculate weighted total score
	score.TotalScore = sa.calculateWeightedScore(score, criteria)

	// Generate selection reason
	score.Reason = sa.generateReason(score, criteria)

	return score
}

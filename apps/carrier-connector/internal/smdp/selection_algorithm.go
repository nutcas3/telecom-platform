package smdp

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

// SelectionCriteria defines parameters for carrier selection
type SelectionCriteria struct {
	Region            string  `json:"region"`
	ProfileType       string  `json:"profile_type"`
	Urgency           string  `json:"urgency"`            // "low", "medium", "high"
	CostSensitivity   float64 `json:"cost_sensitivity"`   // 0.0-1.0
	PerformanceWeight float64 `json:"performance_weight"` // 0.0-1.0
	ReliabilityWeight float64 `json:"reliability_weight"` // 0.0-1.0
}

// CarrierScore represents the scored evaluation of a carrier
type CarrierScore struct {
	CarrierID        string    `json:"carrier_id"`
	Carrier          *Carrier  `json:"carrier"`
	TotalScore       float64   `json:"total_score"`
	PerformanceScore float64   `json:"performance_score"`
	ReliabilityScore float64   `json:"reliability_score"`
	CostScore        float64   `json:"cost_score"`
	RegionScore      float64   `json:"region_score"`
	CapabilityScore  float64   `json:"capability_score"`
	SelectedAt       time.Time `json:"selected_at"`
	Reason           string    `json:"reason"`
}

// SelectionAlgorithm implements intelligent carrier selection
type SelectionAlgorithm struct {
	logger       *logrus.Logger
	history      map[string][]CarrierScore // Selection history per carrier
	maxHistory   int
	learningRate float64
	mlModel      *LearningModel // Machine learning model
}

// NewSelectionAlgorithm creates a new selection algorithm instance
func NewSelectionAlgorithm() *SelectionAlgorithm {
	return &SelectionAlgorithm{
		logger:       logrus.New(),
		history:      make(map[string][]CarrierScore),
		maxHistory:   100,
		learningRate: 0.1,
		mlModel:      NewLearningModel(),
	}
}

// calculateWeightedScore combines all scores with weights
func (sa *SelectionAlgorithm) calculateWeightedScore(score *CarrierScore, criteria *SelectionCriteria) float64 {
	weights := sa.getWeights(criteria)

	total := score.PerformanceScore*weights.performance +
		score.ReliabilityScore*weights.reliability +
		score.CostScore*weights.cost +
		score.RegionScore*weights.region +
		score.CapabilityScore*weights.capability

	return math.Min(100.0, total)
}

// getWeights returns selection weights based on criteria and ML optimization
func (sa *SelectionAlgorithm) getWeights(criteria *SelectionCriteria) struct {
	performance float64
	reliability float64
	cost        float64
	region      float64
	capability  float64
} {
	// Get ML-optimized weights as base
	mlWeights := sa.mlModel.GetOptimizedWeights()

	weights := struct {
		performance float64
		reliability float64
		cost        float64
		region      float64
		capability  float64
	}{
		performance: mlWeights.Performance,
		reliability: mlWeights.Reliability,
		cost:        mlWeights.Cost,
		region:      mlWeights.Region,
		capability:  mlWeights.Capability,
	}

	// Adjust based on urgency (override ML weights for specific business rules)
	switch criteria.Urgency {
	case "high":
		weights.performance = math.Max(weights.performance, 0.4)
		weights.reliability = math.Max(weights.reliability, 0.4)
		weights.cost = math.Min(weights.cost, 0.1)
		weights.region = math.Min(weights.region, 0.05)
		weights.capability = math.Min(weights.capability, 0.05)
	case "low":
		weights.cost = math.Max(weights.cost, 0.4)
		weights.performance = math.Min(weights.performance, 0.2)
		weights.reliability = math.Min(weights.reliability, 0.2)
	}

	// Apply user-defined weights (business preferences override ML)
	if criteria.PerformanceWeight > 0 {
		weights.performance = criteria.PerformanceWeight
	}
	if criteria.ReliabilityWeight > 0 {
		weights.reliability = criteria.ReliabilityWeight
	}
	if criteria.CostSensitivity > 0 {
		weights.cost = criteria.CostSensitivity
	}

	// Normalize weights
	total := weights.performance + weights.reliability + weights.cost + weights.region + weights.capability
	if total > 0 {
		weights.performance /= total
		weights.reliability /= total
		weights.cost /= total
		weights.region /= total
		weights.capability /= total
	}

	return weights
}

// recordSelection records the selection for learning
func (sa *SelectionAlgorithm) recordSelection(score *CarrierScore) {
	history := sa.history[score.CarrierID]
	history = append(history, *score)

	// Limit history size
	if len(history) > sa.maxHistory {
		history = history[1:]
	}

	sa.history[score.CarrierID] = history
}

// GetSelectionHistory returns selection history for a carrier
func (sa *SelectionAlgorithm) GetSelectionHistory(carrierID string) []CarrierScore {
	return sa.history[carrierID]
}

// UpdateLearning updates the algorithm based on feedback using ML model
func (sa *SelectionAlgorithm) UpdateLearning(carrierID string, actualPerformance float64) {
	history := sa.history[carrierID]
	if len(history) == 0 {
		return
	}

	// Get the most recent selection to get expected score
	lastSelection := history[len(history)-1]
	expectedScore := lastSelection.TotalScore

	// Update ML model with performance feedback
	sa.mlModel.UpdateLearning(carrierID, actualPerformance, expectedScore)

	// Legacy learning rate adjustment for backward compatibility
	if actualPerformance < 50.0 {
		sa.learningRate *= 0.95 // Reduce learning rate for poor performance
	} else {
		sa.learningRate *= 1.05 // Increase learning rate for good performance
	}
}

// GetLearningStats returns statistics about the ML learning model
func (sa *SelectionAlgorithm) GetLearningStats() map[string]any {
	return sa.mlModel.GetLearningStats()
}

// PredictPerformance predicts expected performance for a carrier
func (sa *SelectionAlgorithm) PredictPerformance(carrierID string, criteria *SelectionCriteria) float64 {
	return sa.mlModel.PredictPerformance(carrierID, criteria)
}

// GetCarrierPerformance returns detailed performance metrics for a carrier
func (sa *SelectionAlgorithm) GetCarrierPerformance(carrierID string) *PerformanceMetrics {
	return sa.mlModel.GetCarrierPerformance(carrierID)
}

// getHighestPriorityCarrier returns the carrier with the highest priority
func (sa *SelectionAlgorithm) getHighestPriorityCarrier(carriers []*Carrier) *Carrier {
	if len(carriers) == 0 {
		return nil
	}

	highestPriority := carriers[0]
	for _, carrier := range carriers {
		if carrier.Priority > highestPriority.Priority {
			highestPriority = carrier
		}
	}

	return highestPriority
}

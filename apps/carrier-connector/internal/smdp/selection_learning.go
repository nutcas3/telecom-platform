package smdp

import (
	"math"
	"time"
)

// LearningModel represents the machine learning model for carrier selection
type LearningModel struct {
	// Performance tracking
	carrierPerformance  map[string]*PerformanceMetrics
	learningRate        float64
	adaptationThreshold float64

	// Weight optimization
	baseWeights      WeightVector
	optimizedWeights WeightVector

	// Historical data
	performanceHistory map[string][]float64
	selectionHistory   map[string][]SelectionRecord

	// Model parameters
	momentum       float64
	regularization float64
	decayRate      float64
}

// PerformanceMetrics tracks carrier performance over time
type PerformanceMetrics struct {
	SuccessRate         float64
	AverageResponseTime float64
	Reliability         float64
	CostEfficiency      float64
	LastUpdated         time.Time
	SampleCount         int
}

// WeightVector represents the selection weights
type WeightVector struct {
	Performance float64
	Reliability float64
	Cost        float64
	Region      float64
	Capability  float64
}

// SelectionRecord tracks individual selection outcomes
type SelectionRecord struct {
	CarrierID       string
	SelectedAt      time.Time
	ExpectedScore   float64
	ActualScore     float64
	Performance     float64
	ResponseTime    time.Duration
	Success         bool
	SelectionReason string
}

// NewLearningModel creates a new machine learning model
func NewLearningModel() *LearningModel {
	return &LearningModel{
		carrierPerformance:  make(map[string]*PerformanceMetrics),
		performanceHistory:  make(map[string][]float64),
		selectionHistory:    make(map[string][]SelectionRecord),
		learningRate:        0.01,
		adaptationThreshold: 0.05,
		momentum:            0.9,
		regularization:      0.001,
		decayRate:           0.995,
		baseWeights: WeightVector{
			Performance: 0.3,
			Reliability: 0.3,
			Cost:        0.2,
			Region:      0.1,
			Capability:  0.1,
		},
		optimizedWeights: WeightVector{
			Performance: 0.3,
			Reliability: 0.3,
			Cost:        0.2,
			Region:      0.1,
			Capability:  0.1,
		},
	}
}

// UpdateLearning updates the model with performance feedback using ML techniques
func (lm *LearningModel) UpdateLearning(carrierID string, actualPerformance float64, expectedScore float64) {
	// Record the performance outcome
	record := SelectionRecord{
		CarrierID:     carrierID,
		SelectedAt:    time.Now(),
		ExpectedScore: expectedScore,
		ActualScore:   actualPerformance,
		Performance:   actualPerformance,
		Success:       actualPerformance >= 70.0, // 70% threshold for success
	}

	// Add to selection history
	history := lm.selectionHistory[carrierID]
	history = append(history, record)

	// Keep only last 100 records per carrier
	if len(history) > 100 {
		history = history[1:]
	}
	lm.selectionHistory[carrierID] = history

	// Update carrier performance metrics
	lm.updatePerformanceMetrics(carrierID, record)

	// Calculate performance prediction error
	predictionError := actualPerformance - expectedScore

	// If error is significant, trigger weight optimization
	if math.Abs(predictionError) > lm.adaptationThreshold {
		lm.optimizeWeights(carrierID, predictionError)
	}

	// Apply learning rate decay
	lm.learningRate *= lm.decayRate
	if lm.learningRate < 0.001 {
		lm.learningRate = 0.001 // Minimum learning rate
	}
}

// updatePerformanceMetrics updates the performance metrics for a carrier
func (lm *LearningModel) updatePerformanceMetrics(carrierID string, record SelectionRecord) {
	metrics, exists := lm.carrierPerformance[carrierID]
	if !exists {
		metrics = &PerformanceMetrics{
			LastUpdated: time.Now(),
		}
		lm.carrierPerformance[carrierID] = metrics
	}

	// Update metrics with exponential moving average
	alpha := 0.1 // Smoothing factor

	if metrics.SampleCount == 0 {
		// First sample
		metrics.SuccessRate = 1.0
		if !record.Success {
			metrics.SuccessRate = 0.0
		}
		metrics.AverageResponseTime = float64(record.ResponseTime.Milliseconds())
		metrics.Reliability = record.Performance / 100.0
	} else {
		// Exponential moving average update
		if record.Success {
			metrics.SuccessRate = alpha*1.0 + (1-alpha)*metrics.SuccessRate
		} else {
			metrics.SuccessRate = alpha*0.0 + (1-alpha)*metrics.SuccessRate
		}

		responseTimeMs := float64(record.ResponseTime.Milliseconds())
		metrics.AverageResponseTime = alpha*responseTimeMs + (1-alpha)*metrics.AverageResponseTime
		metrics.Reliability = alpha*(record.Performance/100.0) + (1-alpha)*metrics.Reliability
	}

	metrics.CostEfficiency = 1.0 - (float64(record.ResponseTime.Milliseconds()) / 1000.0) // Simple cost proxy
	metrics.LastUpdated = time.Now()
	metrics.SampleCount++
}

// optimizeWeights uses gradient descent to optimize selection weights
func (lm *LearningModel) optimizeWeights(carrierID string, error float64) {
	// Calculate gradients based on performance error
	gradients := lm.calculateGradients(carrierID, error)

	// Apply gradient descent with momentum
	learningRate := lm.learningRate

	// Update optimized weights
	lm.optimizedWeights.Performance -= learningRate * gradients.Performance
	lm.optimizedWeights.Reliability -= learningRate * gradients.Reliability
	lm.optimizedWeights.Cost -= learningRate * gradients.Cost
	lm.optimizedWeights.Region -= learningRate * gradients.Region
	lm.optimizedWeights.Capability -= learningRate * gradients.Capability

	// Apply regularization to prevent overfitting
	lm.applyRegularization()

	// Ensure weights sum to 1 and are non-negative
	lm.normalizeWeights()
}

// calculateGradients calculates the gradient for each weight based on error
func (lm *LearningModel) calculateGradients(carrierID string, error float64) WeightVector {
	metrics := lm.carrierPerformance[carrierID]
	if metrics == nil {
		return WeightVector{}
	}

	// Calculate gradients based on how each factor contributed to the error
	gradients := WeightVector{}

	// Performance gradient: higher error suggests need to adjust performance weight
	gradients.Performance = error * (metrics.Reliability - 0.5) * 2.0

	// Reliability gradient: adjust based on success rate
	gradients.Reliability = error * (metrics.SuccessRate - 0.5) * 2.0

	// Cost gradient: adjust based on cost efficiency
	gradients.Cost = error * (metrics.CostEfficiency - 0.5) * 2.0

	// Region and capability gradients (simplified)
	gradients.Region = error * 0.1
	gradients.Capability = error * 0.1

	return gradients
}

// applyRegularization applies L2 regularization to prevent overfitting
func (lm *LearningModel) applyRegularization() {
	regFactor := lm.regularization

	lm.optimizedWeights.Performance *= (1.0 - regFactor)
	lm.optimizedWeights.Reliability *= (1.0 - regFactor)
	lm.optimizedWeights.Cost *= (1.0 - regFactor)
	lm.optimizedWeights.Region *= (1.0 - regFactor)
	lm.optimizedWeights.Capability *= (1.0 - regFactor)
}

// normalizeWeights ensures weights sum to 1 and are non-negative
func (lm *LearningModel) normalizeWeights() {
	// Ensure non-negative
	if lm.optimizedWeights.Performance < 0 {
		lm.optimizedWeights.Performance = 0
	}
	if lm.optimizedWeights.Reliability < 0 {
		lm.optimizedWeights.Reliability = 0
	}
	if lm.optimizedWeights.Cost < 0 {
		lm.optimizedWeights.Cost = 0
	}
	if lm.optimizedWeights.Region < 0 {
		lm.optimizedWeights.Region = 0
	}
	if lm.optimizedWeights.Capability < 0 {
		lm.optimizedWeights.Capability = 0
	}

	// Normalize to sum to 1
	total := lm.optimizedWeights.Performance + lm.optimizedWeights.Reliability +
		lm.optimizedWeights.Cost + lm.optimizedWeights.Region + lm.optimizedWeights.Capability

	if total > 0 {
		lm.optimizedWeights.Performance /= total
		lm.optimizedWeights.Reliability /= total
		lm.optimizedWeights.Cost /= total
		lm.optimizedWeights.Region /= total
		lm.optimizedWeights.Capability /= total
	} else {
		// Fallback to base weights if all weights become zero
		lm.optimizedWeights = lm.baseWeights
	}
}

// GetOptimizedWeights returns the current optimized weights for selection
func (lm *LearningModel) GetOptimizedWeights() WeightVector {
	return lm.optimizedWeights
}

// GetCarrierPerformance returns performance metrics for a carrier
func (lm *LearningModel) GetCarrierPerformance(carrierID string) *PerformanceMetrics {
	return lm.carrierPerformance[carrierID]
}

// PredictPerformance predicts the expected performance for a carrier
func (lm *LearningModel) PredictPerformance(carrierID string, criteria *SelectionCriteria) float64 {
	metrics := lm.carrierPerformance[carrierID]
	if metrics == nil {
		return 75.0 // Default prediction for new carriers
	}

	// Use weighted combination of performance metrics
	weights := lm.optimizedWeights

	prediction := metrics.SuccessRate*weights.Reliability*100 +
		(1.0-metrics.AverageResponseTime/1000.0)*weights.Performance*100 +
		metrics.CostEfficiency*weights.Cost*100 +
		0.5*weights.Region*100 + // Simplified region prediction
		0.5*weights.Capability*100 // Simplified capability prediction

	return math.Min(100.0, math.Max(0.0, prediction))
}

// GetLearningStats returns statistics about the learning model
func (lm *LearningModel) GetLearningStats() map[string]any {
	stats := make(map[string]any)

	stats["learning_rate"] = lm.learningRate
	stats["total_carriers_tracked"] = len(lm.carrierPerformance)
	stats["optimized_weights"] = lm.optimizedWeights
	stats["base_weights"] = lm.baseWeights

	// Calculate overall performance improvement
	totalImprovement := 0.0
	sampleCount := 0

	for _, history := range lm.selectionHistory {
		if len(history) > 10 { // Only consider carriers with sufficient data
			firstHalf := history[:len(history)/2]
			secondHalf := history[len(history)/2:]

			firstAvg := 0.0
			for _, record := range firstHalf {
				firstAvg += record.ActualScore
			}
			firstAvg /= float64(len(firstHalf))

			secondAvg := 0.0
			for _, record := range secondHalf {
				secondAvg += record.ActualScore
			}
			secondAvg /= float64(len(secondHalf))

			improvement := secondAvg - firstAvg
			totalImprovement += improvement
			sampleCount++
		}
	}

	if sampleCount > 0 {
		stats["average_performance_improvement"] = totalImprovement / float64(sampleCount)
	}

	return stats
}

package smdp

import (
	"context"
	"testing"
)

// TestMachineLearningModel tests the ML learning capabilities
func TestMachineLearningModel(t *testing.T) {
	mlModel := NewLearningModel()

	// Test initial state
	t.Run("InitialState", func(t *testing.T) {
		weights := mlModel.GetOptimizedWeights()

		// Should start with base weights
		if weights.Performance < 0.29 || weights.Performance > 0.31 {
			t.Errorf("Expected performance weight around 0.3, got %.3f", weights.Performance)
		}

		stats := mlModel.GetLearningStats()
		if stats["learning_rate"].(float64) <= 0 {
			t.Error("Learning rate should be positive")
		}
	})

	// Test learning updates
	t.Run("LearningUpdates", func(t *testing.T) {
		carrierID := "test-ml-carrier"

		// Simulate multiple selections with feedback
		for i := range 10 {
			expectedScore := float64(70 + i*2) // Improving expected scores
			actualScore := float64(65 + i*3)   // Actual performance

			mlModel.UpdateLearning(carrierID, actualScore, expectedScore)
		}

		// Check that weights have been optimized
		optimizedWeights := mlModel.GetOptimizedWeights()

		// Weights should have changed from original defaults (0.3, 0.3, 0.2, 0.1, 0.1)
		if optimizedWeights.Performance == 0.3 && optimizedWeights.Reliability == 0.3 {
			t.Error("Expected weights to change after learning")
		}

		// Weights should still sum to 1
		total := optimizedWeights.Performance + optimizedWeights.Reliability +
			optimizedWeights.Cost + optimizedWeights.Region + optimizedWeights.Capability
		if total < 0.99 || total > 1.01 {
			t.Errorf("Weights should sum to 1.0, got %.3f", total)
		}
	})

	// Test performance metrics tracking
	t.Run("PerformanceMetrics", func(t *testing.T) {
		carrierID := "metrics-carrier"

		// Add some performance data
		mlModel.UpdateLearning(carrierID, 85.0, 80.0)
		mlModel.UpdateLearning(carrierID, 90.0, 85.0)
		mlModel.UpdateLearning(carrierID, 78.0, 82.0)

		metrics := mlModel.GetCarrierPerformance(carrierID)
		if metrics == nil {
			t.Fatal("Expected performance metrics for carrier")
		}

		if metrics.SampleCount != 3 {
			t.Errorf("Expected 3 samples, got %d", metrics.SampleCount)
		}

		if metrics.SuccessRate < 0 || metrics.SuccessRate > 1 {
			t.Errorf("Success rate should be between 0 and 1, got %.3f", metrics.SuccessRate)
		}
	})

	// Test performance prediction
	t.Run("PerformancePrediction", func(t *testing.T) {
		carrierID := "prediction-carrier"
		criteria := &SelectionCriteria{
			Region:            "US",
			ProfileType:       "operational",
			Urgency:           "medium",
			CostSensitivity:   0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		// Add training data
		for i := range 5 {
			mlModel.UpdateLearning(carrierID, 80.0+float64(i*2), 75.0+float64(i))
		}

		prediction := mlModel.PredictPerformance(carrierID, criteria)
		if prediction < 0 || prediction > 100 {
			t.Errorf("Prediction should be between 0 and 100, got %.2f", prediction)
		}

		// Prediction for unknown carrier should return default
		unknownPrediction := mlModel.PredictPerformance("unknown-carrier", criteria)
		if unknownPrediction != 75.0 {
			t.Errorf("Expected default prediction for unknown carrier, got %.2f", unknownPrediction)
		}
	})
}

// TestMLIntegration tests the integration of ML with SelectionAlgorithm
func TestMLIntegration(t *testing.T) {
	selector := NewSelectionAlgorithm()

	// Create test carriers
	carriers := []*Carrier{
		createTestCarrier("ml-test-1", "ML Test 1", "US", 80, 85, 150, 5.0),
		createTestCarrier("ml-test-2", "ML Test 2", "US", 75, 90, 120, 4.0),
	}

	criteria := &SelectionCriteria{
		Region:            "US",
		ProfileType:       "operational",
		Urgency:           "medium",
		CostSensitivity:   0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	// Test initial selection
	t.Run("InitialSelection", func(t *testing.T) {
		score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}

		// Provide learning feedback
		selector.UpdateLearning(score.CarrierID, 85.0)

		// Check learning stats
		stats := selector.GetLearningStats()
		if stats["total_carriers_tracked"].(int) == 0 {
			t.Error("Expected carriers to be tracked in ML model")
		}
	})

	// Test learning over multiple iterations
	t.Run("IterativeLearning", func(t *testing.T) {
		weightsBefore := selector.mlModel.GetOptimizedWeights()

		// Perform multiple selections with feedback
		for i := range 5 {
			score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
			if err != nil {
				t.Fatalf("Selection %d failed: %v", i, err)
			}

			// Simulate performance feedback (varying performance)
			actualPerformance := 70.0 + float64(i*3)
			selector.UpdateLearning(score.CarrierID, actualPerformance)
		}

		weightsAfter := selector.mlModel.GetOptimizedWeights()

		// Weights should have been optimized
		if weightsAfter.Performance == weightsBefore.Performance &&
			weightsAfter.Reliability == weightsBefore.Reliability {
			t.Error("Expected weights to change after learning iterations")
		}

		// Check performance prediction
		prediction := selector.PredictPerformance("ml-test-1", criteria)
		if prediction < 0 || prediction > 100 {
			t.Errorf("Prediction should be between 0 and 100, got %.2f", prediction)
		}
	})

	// Test carrier performance metrics
	t.Run("CarrierPerformanceMetrics", func(t *testing.T) {
		// Get performance metrics for a carrier
		metrics := selector.GetCarrierPerformance("ml-test-1")
		if metrics == nil {
			t.Error("Expected performance metrics for carrier")
		}

		if metrics.SampleCount == 0 {
			t.Error("Expected sample count > 0 after learning")
		}
	})
}

// TestMLWeightOptimization tests the weight optimization logic
func TestMLWeightOptimization(t *testing.T) {
	mlModel := NewLearningModel()

	t.Run("GradientDescent", func(t *testing.T) {
		carrierID := "gradient-test"

		// Simulate consistent poor performance to trigger weight optimization
		for range 10 {
			expectedScore := 80.0
			actualScore := 50.0 // Consistently poor performance

			mlModel.UpdateLearning(carrierID, actualScore, expectedScore)
		}

		weights := mlModel.GetOptimizedWeights()

		// After consistent poor performance, weights should be adjusted
		// This is a simplified test - in practice, the optimization is more complex
		total := weights.Performance + weights.Reliability + weights.Cost + weights.Region + weights.Capability
		if total < 0.99 || total > 1.01 {
			t.Errorf("Weights should still sum to 1.0 after optimization, got %.3f", total)
		}

		// All weights should be non-negative
		if weights.Performance < 0 || weights.Reliability < 0 || weights.Cost < 0 ||
			weights.Region < 0 || weights.Capability < 0 {
			t.Error("All weights should be non-negative")
		}
	})

	t.Run("Regularization", func(t *testing.T) {
		// Test that regularization prevents extreme weight values
		carrierID := "regularization-test"

		// Simulate extreme performance variations
		for i := range 20 {
			expectedScore := 80.0
			actualScore := 20.0 + float64(i*3) // Wide range of performance

			mlModel.UpdateLearning(carrierID, actualScore, expectedScore)
		}

		weights := mlModel.GetOptimizedWeights()

		// Weights should be reasonable (not too extreme)
		for _, weight := range []float64{
			weights.Performance, weights.Reliability, weights.Cost,
			weights.Region, weights.Capability} {
			if weight < 0.05 || weight > 0.7 {
				t.Errorf("Weight %.3f seems too extreme (should be between 0.05 and 0.7)", weight)
			}
		}
	})
}

// TestMLLearningRate tests the learning rate adaptation
func TestMLLearningRate(t *testing.T) {
	mlModel := NewLearningModel()

	initialLearningRate := mlModel.GetLearningStats()["learning_rate"].(float64)

	// Add some learning data
	for range 10 {
		mlModel.UpdateLearning("rate-test", 75.0, 70.0)
	}

	finalLearningRate := mlModel.GetLearningStats()["learning_rate"].(float64)

	// Learning rate should have decayed
	if finalLearningRate >= initialLearningRate {
		t.Error("Learning rate should decay over time")
	}

	// Learning rate should not go below minimum
	if finalLearningRate < 0.001 {
		t.Error("Learning rate should not go below minimum threshold")
	}
}

// BenchmarkMLModel benchmarks the ML model performance
func BenchmarkMLModel(b *testing.B) {
	mlModel := NewLearningModel()
	carrierID := "benchmark-carrier"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mlModel.UpdateLearning(carrierID, 75.0, 70.0)
	}
}

// BenchmarkSelectionWithML benchmarks selection with ML integration
func BenchmarkSelectionWithML(b *testing.B) {
	selector := NewSelectionAlgorithm()

	carriers := []*Carrier{
		createTestCarrier("benchmark-1", "Benchmark 1", "US", 80, 85, 150, 5.0),
		createTestCarrier("benchmark-2", "Benchmark 2", "US", 75, 90, 120, 4.0),
	}

	criteria := &SelectionCriteria{
		Region:            "US",
		ProfileType:       "operational",
		Urgency:           "medium",
		CostSensitivity:   0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			b.Fatalf("Selection failed: %v", err)
		}
		selector.UpdateLearning("benchmark-1", 75.0)
	}
}

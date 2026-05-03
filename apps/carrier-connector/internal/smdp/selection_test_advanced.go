package smdp

import (
	"context"
	"testing"
)

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	selector := NewSelectionAlgorithm()

	// Test empty carrier list
	t.Run("EmptyCarrierList", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Region:            "US",
			ProfileType:       "operational",
			Urgency:           "medium",
			CostSensitivity:   0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		_, err := selector.SelectOptimalCarrier(context.Background(), []*Carrier{}, criteria)
		if err == nil {
			t.Error("Expected error for empty carrier list")
		}
	})

	// Test all unhealthy carriers
	t.Run("AllUnhealthyCarriers", func(t *testing.T) {
		unhealthyCarriers := []*Carrier{
			createUnhealthyCarrier("broken1", "Broken 1", "US"),
			createUnhealthyCarrier("broken2", "Broken 2", "US"),
		}

		criteria := &SelectionCriteria{
			Region:            "US",
			ProfileType:       "operational",
			Urgency:           "medium",
			CostSensitivity:   0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		_, err := selector.SelectOptimalCarrier(context.Background(), unhealthyCarriers, criteria)
		if err == nil {
			t.Error("Expected error when all carriers are unhealthy")
		}
	})

	// Test inactive carriers
	t.Run("InactiveCarriers", func(t *testing.T) {
		inactiveCarriers := []*Carrier{
			createInactiveCarrier("inactive1", "Inactive 1", "US"),
			createInactiveCarrier("inactive2", "Inactive 2", "US"),
		}

		criteria := &SelectionCriteria{
			Region:            "US",
			ProfileType:       "operational",
			Urgency:           "medium",
			CostSensitivity:   0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		_, err := selector.SelectOptimalCarrier(context.Background(), inactiveCarriers, criteria)
		if err == nil {
			t.Error("Expected error when all carriers are inactive")
		}
	})
}

// TestMultipleSelections tests behavior with multiple selections
func TestMultipleSelections(t *testing.T) {
	selector := NewSelectionAlgorithm()

	carriers := []*Carrier{
		createTestCarrier("carrier1", "Carrier 1", "US", 90, 95, 100, 5.0),
		createTestCarrier("carrier2", "Carrier 2", "US", 85, 98, 120, 4.0),
		createTestCarrier("carrier3", "Carrier 3", "US", 80, 97, 110, 6.0),
	}

	criteria := &SelectionCriteria{
		Region:            "US",
		ProfileType:       "operational",
		Urgency:           "medium",
		CostSensitivity:   0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	// Perform multiple selections
	selections := make(map[string]int)
	for i := range 20 {
		score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Selection %d failed: %v", i, err)
		}
		selections[score.CarrierID]++
	}

	// Check that we have some distribution
	totalSelections := 0
	for _, count := range selections {
		totalSelections += count
	}
	if totalSelections != 20 {
		t.Errorf("Expected 20 total selections, got %d", totalSelections)
	}

	t.Logf("Selection distribution: %+v", selections)
}

// TestLearningFeedback tests the learning feedback mechanism
func TestLearningFeedback(t *testing.T) {
	selector := NewSelectionAlgorithm()

	carriers := []*Carrier{
		createTestCarrier("learning-carrier", "Learning Carrier", "US", 80, 90, 150, 5.0),
	}

	criteria := &SelectionCriteria{
		Region:            "US",
		ProfileType:       "operational",
		Urgency:           "medium",
		CostSensitivity:   0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	// Initial selection
	_, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
	if err != nil {
		t.Fatalf("Initial selection failed: %v", err)
	}

	initialLearningRate := selector.learningRate

	// Update with poor performance
	selector.UpdateLearning("learning-carrier", 30.0)
	learningRateAfterPoor := selector.learningRate

	// Update with good performance
	selector.UpdateLearning("learning-carrier", 90.0)
	learningRateAfterGood := selector.learningRate

	// Verify learning rate changes
	if learningRateAfterPoor >= initialLearningRate {
		t.Error("Learning rate should decrease after poor performance")
	}

	if learningRateAfterGood <= learningRateAfterPoor {
		t.Error("Learning rate should increase after good performance")
	}

	t.Logf("Learning rate progression: %.4f -> %.4f -> %.4f",
		initialLearningRate, learningRateAfterPoor, learningRateAfterGood)
}

// Helper function to create inactive carrier
func createInactiveCarrier(id, name, country string) *Carrier {
	carrier := createTestCarrier(id, name, country, 50, 50, 500, 1.0)
	carrier.IsActive = false
	return carrier
}

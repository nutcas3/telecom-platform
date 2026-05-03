package smdp

import (
	"context"
	"testing"
)

// TestScoringComponents tests individual scoring components
func TestScoringComponents(t *testing.T) {
	selector := NewSelectionAlgorithm()

	// Test performance scoring
	t.Run("PerformanceScoring", func(t *testing.T) {
		carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)
		score := selector.calculatePerformanceScore(carrier)

		if score < 0 || score > 100 {
			t.Errorf("Performance score should be between 0 and 100, got %.2f", score)
		}

		t.Logf("Performance score: %.2f", score)
	})

	// Test reliability scoring
	t.Run("ReliabilityScoring", func(t *testing.T) {
		carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)
		score := selector.calculateReliabilityScore(carrier)

		if score < 0 || score > 100 {
			t.Errorf("Reliability score should be between 0 and 100, got %.2f", score)
		}

		t.Logf("Reliability score: %.2f", score)
	})

	// Test cost scoring
	t.Run("CostScoring", func(t *testing.T) {
		carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)
		criteria := &SelectionCriteria{CostSensitivity: 0.5}
		score := selector.calculateCostScore(carrier, criteria)

		if score < 0 || score > 100 {
			t.Errorf("Cost score should be between 0 and 100, got %.2f", score)
		}

		t.Logf("Cost score: %.2f", score)
	})

	// Test region scoring
	t.Run("RegionScoring", func(t *testing.T) {
		carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)

		// Test matching region
		score1 := selector.calculateRegionScore(carrier, "US")
		if score1 != 100.0 {
			t.Errorf("Expected perfect score for matching region, got %.2f", score1)
		}

		// Test non-matching region
		score2 := selector.calculateRegionScore(carrier, "DE")
		if score2 != 50.0 {
			t.Errorf("Expected neutral score for non-matching region, got %.2f", score2)
		}

		// Test empty region
		score3 := selector.calculateRegionScore(carrier, "")
		if score3 != 50.0 {
			t.Errorf("Expected neutral score for empty region, got %.2f", score3)
		}

		t.Logf("Region scores - Match: %.2f, Non-match: %.2f, Empty: %.2f", score1, score2, score3)
	})

	// Test capability scoring
	t.Run("CapabilityScoring", func(t *testing.T) {
		carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)

		// Test supported profile type
		score1 := selector.calculateCapabilityScore(carrier, "operational")
		if score1 < 80.0 {
			t.Errorf("Expected high score for supported profile type, got %.2f", score1)
		}

		// Test unsupported profile type
		score2 := selector.calculateCapabilityScore(carrier, "unsupported")
		if score2 < 50.0 {
			t.Errorf("Expected base score for unsupported profile type, got %.2f", score2)
		}

		t.Logf("Capability scores - Supported: %.2f, Unsupported: %.2f", score1, score2)
	})
}

// TestWeightCalculation tests the weight calculation logic
func TestWeightCalculation(t *testing.T) {
	selector := NewSelectionAlgorithm()

	// Test default weights
	t.Run("DefaultWeights", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Urgency: "medium",
		}
		weights := selector.getWeights(criteria)

		total := weights.performance + weights.reliability + weights.cost + weights.region + weights.capability
		if total < 0.99 || total > 1.01 { // Allow for floating point precision
			t.Errorf("Weights should sum to 1.0, got %.2f", total)
		}

		t.Logf("Default weights - P: %.2f, R: %.2f, C: %.2f, Reg: %.2f, Cap: %.2f",
			weights.performance, weights.reliability, weights.cost, weights.region, weights.capability)
	})

	// Test high urgency weights
	t.Run("HighUrgencyWeights", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Urgency: "high",
		}
		weights := selector.getWeights(criteria)

		if weights.performance < 0.39 || weights.reliability < 0.39 {
			t.Errorf("High urgency should prioritize performance and reliability")
		}

		t.Logf("High urgency weights - P: %.2f, R: %.2f, C: %.2f, Reg: %.2f, Cap: %.2f",
			weights.performance, weights.reliability, weights.cost, weights.region, weights.capability)
	})

	// Test low urgency weights
	t.Run("LowUrgencyWeights", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Urgency: "low",
		}
		weights := selector.getWeights(criteria)

		if weights.cost < 0.39 {
			t.Errorf("Low urgency should prioritize cost")
		}

		t.Logf("Low urgency weights - P: %.2f, R: %.2f, C: %.2f, Reg: %.2f, Cap: %.2f",
			weights.performance, weights.reliability, weights.cost, weights.region, weights.capability)
	})

	// Test custom weights
	t.Run("CustomWeights", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Urgency:           "medium",
			PerformanceWeight: 0.7,
			ReliabilityWeight: 0.2,
			CostSensitivity:   0.1,
		}
		weights := selector.getWeights(criteria)

		total := weights.performance + weights.reliability + weights.cost + weights.region + weights.capability
		if total < 0.99 || total > 1.01 {
			t.Errorf("Custom weights should sum to 1.0, got %.2f", total)
		}

		t.Logf("Custom weights - P: %.2f, R: %.2f, C: %.2f, Reg: %.2f, Cap: %.2f",
			weights.performance, weights.reliability, weights.cost, weights.region, weights.capability)
	})
}

// TestSelectionHistory tests the selection history functionality
func TestSelectionHistory(t *testing.T) {
	selector := NewSelectionAlgorithm()

	carriers := []*Carrier{
		createTestCarrier("test-carrier", "Test Carrier", "US", 80, 90, 150, 5.0),
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
	for i := range 5 {
		_, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Failed to select optimal carrier on iteration %d: %v", i, err)
		}
	}

	// Check selection history
	history := selector.GetSelectionHistory("test-carrier")
	if len(history) != 5 {
		t.Errorf("Expected 5 selections in history, got %d", len(history))
	}

	// Test learning feedback
	selector.UpdateLearning("test-carrier", 85.0)

	t.Logf("Selection history contains %d entries", len(history))
}

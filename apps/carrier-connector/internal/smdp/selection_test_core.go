package smdp

import (
	"context"
	"testing"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
)

// TestSelectionAlgorithm tests the intelligent carrier selection algorithm
func TestSelectionAlgorithm(t *testing.T) {
	// Create selection algorithm
	selector := NewSelectionAlgorithm()

	// Create test carriers with different characteristics
	carriers := []*Carrier{
		createTestCarrier("att-us", "AT&T US", "US", 90, 98, 150, 10.5),
		createTestCarrier("verizon-us", "Verizon US", "US", 85, 99, 120, 8.2),
		createTestCarrier("tmobile-de", "T-Mobile DE", "DE", 75, 95, 200, 6.8),
		createTestCarrier("orange-fr", "Orange FR", "FR", 70, 95, 180, 4.5),
	}

	// Test case 1: High priority US request
	t.Run("HighPriorityUSRequest", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Region:           "US",
			ProfileType:      "operational",
			Urgency:          "high",
			CostSensitivity:  0.2,
			PerformanceWeight: 0.6,
			ReliabilityWeight: 0.6,
		}

		score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Failed to select optimal carrier: %v", err)
		}

		// Should select a US carrier with high performance
		if score.CarrierID != "att-us" && score.CarrierID != "verizon-us" {
			t.Errorf("Expected US carrier, got %s", score.CarrierID)
		}

		if score.TotalScore < 80 {
			t.Errorf("Expected high score for high priority request, got %.2f", score.TotalScore)
		}

		t.Logf("Selected carrier: %s with score %.2f", score.CarrierID, score.TotalScore)
	})

	// Test case 2: Cost-optimized European request
	t.Run("CostOptimizedEURequest", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Region:           "DE",
			ProfileType:      "operational",
			Urgency:          "low",
			CostSensitivity:  0.8,
			PerformanceWeight: 0.2,
			ReliabilityWeight: 0.3,
		}

		score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Failed to select optimal carrier: %v", err)
		}

		// Should prefer European carrier for better region compatibility
		if score.CarrierID != "tmobile-de" && score.CarrierID != "orange-fr" {
			t.Logf("Note: Selected non-EU carrier %s for DE region", score.CarrierID)
		}

		t.Logf("Selected carrier: %s with score %.2f", score.CarrierID, score.TotalScore)
	})

	// Test case 3: Balanced global request
	t.Run("BalancedGlobalRequest", func(t *testing.T) {
		criteria := &SelectionCriteria{
			Region:           "",
			ProfileType:      "operational",
			Urgency:          "medium",
			CostSensitivity:  0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
		if err != nil {
			t.Fatalf("Failed to select optimal carrier: %v", err)
		}

		// Should select the best overall performer
		if score.TotalScore < 70 {
			t.Errorf("Expected reasonable score for balanced request, got %.2f", score.TotalScore)
		}

		t.Logf("Selected carrier: %s with score %.2f", score.CarrierID, score.TotalScore)
	})

	// Test case 4: No healthy carriers available
	t.Run("NoHealthyCarriers", func(t *testing.T) {
		unhealthyCarriers := []*Carrier{
			createUnhealthyCarrier("broken-carrier", "Broken Carrier", "US"),
		}

		criteria := &SelectionCriteria{
			Region:           "US",
			ProfileType:      "operational",
			Urgency:          "medium",
			CostSensitivity:  0.5,
			PerformanceWeight: 0.4,
			ReliabilityWeight: 0.4,
		}

		_, err := selector.SelectOptimalCarrier(context.Background(), unhealthyCarriers, criteria)
		if err == nil {
			t.Error("Expected error when no healthy carriers available")
		}

		t.Logf("Correctly returned error: %v", err)
	})
}

// TestSelectionLearning tests the learning capabilities
func TestSelectionLearning(t *testing.T) {
	selector := NewSelectionAlgorithm()

	carriers := []*Carrier{
		createTestCarrier("test-carrier", "Test Carrier", "US", 80, 90, 150, 5.0),
	}

	criteria := &SelectionCriteria{
		Region:           "US",
		ProfileType:      "operational",
		Urgency:          "medium",
		CostSensitivity:  0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	// Perform selection
	score, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
	if err != nil {
		t.Fatalf("Failed to select optimal carrier: %v", err)
	}

	// Check selection history
	history := selector.GetSelectionHistory("test-carrier")
	if len(history) != 1 {
		t.Errorf("Expected 1 selection in history, got %d", len(history))
	}

	// Update learning with performance feedback
	selector.UpdateLearning("test-carrier", 85.0)

	// Perform another selection to see if learning affects results
	score2, err := selector.SelectOptimalCarrier(context.Background(), carriers, criteria)
	if err != nil {
		t.Fatalf("Failed to select optimal carrier on second attempt: %v", err)
	}

	t.Logf("First selection score: %.2f, Second selection score: %.2f", 
		score.TotalScore, score2.TotalScore)
}

// Helper functions

func createTestCarrier(id, name, country string, priority, successRate int, responseTime int64, requestRate float64) *Carrier {
	return &Carrier{
		ID:          id,
		Name:        name,
		CountryCode: country,
		MCC:         "310",
		MNC:         "410",
		IsActive:    true,
		Priority:    priority,
		HealthStatus: CarrierStatusHealthy,
		LastHealthCheck: time.Now(),
		ES2Config: &config.ES2Config{
			BaseURL:                  "https://example.com",
			APIKey:                   "test-key",
			InsecureSkipVerify:       false,
			FunctionalityRequesterID: "telecom-platform",
		},
		Capabilities: &CarrierCapabilities{
			SupportedProfileTypes: []string{"operational", "testing"},
			Features:              []string{"bulk_download"},
			MaxConcurrentRequests: 100,
		},
		Metrics: &CarrierMetrics{
			TotalRequests:       uint64(1000 * requestRate),
			SuccessfulRequests:  uint64(float64(1000*requestRate) * float64(successRate) / 100),
			FailedRequests:      uint64(float64(1000*requestRate) * float64(100-successRate) / 100),
			AverageResponseTime: time.Duration(responseTime) * time.Millisecond,
			RequestRate:         requestRate,
		},
	}
}

func createUnhealthyCarrier(id, name, country string) *Carrier {
	carrier := createTestCarrier(id, name, country, 50, 50, 500, 1.0)
	carrier.HealthStatus = CarrierStatusUnhealthy
	return carrier
}

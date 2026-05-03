package smdp

import (
	"context"
	"testing"
)

// BenchmarkSelectionAlgorithm benchmarks the selection algorithm performance
func BenchmarkSelectionAlgorithm(b *testing.B) {
	selector := NewSelectionAlgorithm()

	carriers := make([]*Carrier, 100)
	for i := range 100 {
		carriers[i] = createTestCarrier(
			"carrier-"+string(rune(i)),
			"Carrier "+string(rune(i)),
			"US",
			50+i%50,
			80+i%20,
			100+int64(i*10),
			1.0+float64(i)*0.1,
		)
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
	}
}

// BenchmarkScoring benchmarks individual scoring functions
func BenchmarkScoring(b *testing.B) {
	selector := NewSelectionAlgorithm()
	carrier := createTestCarrier("test", "Test", "US", 80, 95, 150, 5.0)
	criteria := &SelectionCriteria{CostSensitivity: 0.5}

	b.Run("PerformanceScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selector.calculatePerformanceScore(carrier)
		}
	})

	b.Run("ReliabilityScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selector.calculateReliabilityScore(carrier)
		}
	})

	b.Run("CostScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selector.calculateCostScore(carrier, criteria)
		}
	})

	b.Run("RegionScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selector.calculateRegionScore(carrier, "US")
		}
	})

	b.Run("CapabilityScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selector.calculateCapabilityScore(carrier, "operational")
		}
	})
}

// BenchmarkWeightCalculation benchmarks weight calculation
func BenchmarkWeightCalculation(b *testing.B) {
	selector := NewSelectionAlgorithm()

	criteria := &SelectionCriteria{
		Urgency:           "medium",
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.3,
		CostSensitivity:   0.3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selector.getWeights(criteria)
	}
}

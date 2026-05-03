package handlers

import (
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// generateRecommendations generates recommendations based on carrier selection
func (h *SelectionHandler) generateRecommendations(score *smdp.CarrierScore, criteria *smdp.SelectionCriteria) []string {
	recommendations := []string{}

	// Performance-based recommendations
	if score.PerformanceScore > 80 {
		recommendations = append(recommendations, "Excellent performance - suitable for critical operations")
	} else if score.PerformanceScore < 60 {
		recommendations = append(recommendations, "Consider monitoring performance closely")
	}

	// Cost-based recommendations
	if score.CostScore > 80 {
		recommendations = append(recommendations, "Cost-effective choice for budget-conscious operations")
	}

	// Reliability recommendations
	if score.ReliabilityScore > 90 {
		recommendations = append(recommendations, "Highly reliable - suitable for mission-critical applications")
	}

	// Urgency-based recommendations
	if criteria.Urgency == "high" && score.TotalScore < 80 {
		recommendations = append(recommendations, "Consider manual override for high-priority requests")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Carrier appears suitable for requested operation")
	}

	return recommendations
}

// generateHistoryAnalytics generates analytics for selection history
func (h *SelectionHandler) generateHistoryAnalytics(history []smdp.CarrierScore) map[string]any {
	analytics := make(map[string]any)

	if len(history) == 0 {
		analytics["message"] = "No selection history available"
		return analytics
	}

	// Calculate basic statistics
	var totalScore float64
	var performanceSum float64
	var reliabilitySum float64

	for _, score := range history {
		totalScore += score.TotalScore
		performanceSum += score.PerformanceScore
		reliabilitySum += score.ReliabilityScore
	}

	analytics["total_selections"] = len(history)
	analytics["average_score"] = totalScore / float64(len(history))
	analytics["average_performance_score"] = performanceSum / float64(len(history))
	analytics["average_reliability_score"] = reliabilitySum / float64(len(history))

	// Time-based analytics
	if len(history) > 1 {
		firstSelection := history[0].SelectedAt
		lastSelection := history[len(history)-1].SelectedAt
		duration := lastSelection.Sub(firstSelection)
		analytics["selection_span_days"] = int(duration.Hours() / 24)

		// Selection frequency
		frequency := float64(len(history)) / duration.Hours()
		analytics["selections_per_hour"] = frequency
	}

	// Score distribution
	highScores := 0
	mediumScores := 0
	lowScores := 0

	for _, score := range history {
		if score.TotalScore >= 80 {
			highScores++
		} else if score.TotalScore >= 60 {
			mediumScores++
		} else {
			lowScores++
		}
	}

	analytics["score_distribution"] = map[string]int{
		"high":   highScores,
		"medium": mediumScores,
		"low":    lowScores,
	}

	return analytics
}

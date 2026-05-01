package services

import (
	"context"
	"fmt"
	"time"
)

// GetCarrierAnalytics returns comprehensive carrier analytics
func (s *SelectionService) GetCarrierAnalytics(ctx context.Context) (*CarrierAnalyticsResponse, error) {
	s.logger.Info("Generating carrier analytics")

	// Get carrier status
	carriers := s.manager.GetCarrierStatus()

	analytics := &CarrierAnalyticsResponse{
		GeneratedAt:     time.Now(),
		TotalCarriers:   len(carriers),
		HealthyCarriers: 0,
		Analytics:       make(map[string]*CarrierAnalytics),
		Summary:         &AnalyticsSummary{},
	}

	var totalScore float64
	var totalRequests uint64
	var totalSuccess uint64

	for carrierID, carrier := range carriers {
		carrierAnalytics := &CarrierAnalytics{
			CarrierID:   carrierID,
			CarrierName: carrier.Name,
			Region:      carrier.CountryCode,
			Status:      string(carrier.HealthStatus),
			LastCheck:   carrier.LastHealthCheck,
		}

		// Get selection history
		history := s.manager.GetSelectionHistory(carrierID)
		carrierAnalytics.SelectionCount = len(history)

		if len(history) > 0 {
			var scoreSum float64
			for _, score := range history {
				scoreSum += score.TotalScore
			}
			carrierAnalytics.AverageScore = scoreSum / float64(len(history))
			carrierAnalytics.LastSelected = history[len(history)-1].SelectedAt
		}

		// Calculate performance metrics
		if carrier.Metrics.TotalRequests > 0 {
			carrierAnalytics.SuccessRate = float64(carrier.Metrics.SuccessfulRequests) / float64(carrier.Metrics.TotalRequests) * 100
			carrierAnalytics.AverageResponseTime = carrier.Metrics.AverageResponseTime.Milliseconds()
			carrierAnalytics.RequestRate = carrier.Metrics.RequestRate
		}

		// Update totals
		totalScore += carrierAnalytics.AverageScore
		totalRequests += carrier.Metrics.TotalRequests
		totalSuccess += carrier.Metrics.SuccessfulRequests

		// Count healthy carriers
		if carrier.HealthStatus == "healthy" {
			analytics.HealthyCarriers++
		}

		// Generate carrier-specific recommendations
		carrierAnalytics.Recommendations = s.generateCarrierRecommendations(carrierAnalytics)

		analytics.Analytics[carrierID] = carrierAnalytics
	}

	// Calculate summary metrics
	if len(carriers) > 0 {
		analytics.Summary.AverageScore = totalScore / float64(len(carriers))
	}
	if totalRequests > 0 {
		analytics.Summary.OverallSuccessRate = float64(totalSuccess) / float64(totalRequests) * 100
	}

	analytics.Summary.TotalRequests = totalRequests
	analytics.Summary.OverallHealth = float64(analytics.HealthyCarriers) / float64(len(carriers)) * 100

	return analytics, nil
}

// CarrierAnalyticsResponse represents the response for carrier analytics
type CarrierAnalyticsResponse struct {
	GeneratedAt     time.Time                    `json:"generated_at"`
	TotalCarriers   int                          `json:"total_carriers"`
	HealthyCarriers int                          `json:"healthy_carriers"`
	Analytics       map[string]*CarrierAnalytics `json:"analytics"`
	Summary         *AnalyticsSummary            `json:"summary"`
}

// CarrierAnalytics represents analytics for a single carrier
type CarrierAnalytics struct {
	CarrierID           string    `json:"carrier_id"`
	CarrierName         string    `json:"carrier_name"`
	Region              string    `json:"region"`
	Status              string    `json:"status"`
	SelectionCount      int       `json:"selection_count"`
	AverageScore        float64   `json:"average_score"`
	SuccessRate         float64   `json:"success_rate"`
	AverageResponseTime int64     `json:"average_response_time_ms"`
	RequestRate         float64   `json:"request_rate"`
	LastSelected        time.Time `json:"last_selected"`
	LastCheck           time.Time `json:"last_check"`
	Recommendations     []string  `json:"recommendations"`
}

// AnalyticsSummary represents overall analytics summary
type AnalyticsSummary struct {
	AverageScore       float64 `json:"average_score"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	TotalRequests      uint64  `json:"total_requests"`
	OverallHealth      float64 `json:"overall_health"`
}

// generateCarrierRecommendations generates recommendations for a specific carrier
func (s *SelectionService) generateCarrierRecommendations(analytics *CarrierAnalytics) []string {
	recommendations := []string{}

	// Performance recommendations
	if analytics.SuccessRate < 95 {
		recommendations = append(recommendations, "Monitor success rate - below optimal threshold")
	}
	if analytics.AverageResponseTime > 500 {
		recommendations = append(recommendations, "Consider optimizing response time")
	}
	if analytics.RequestRate > 100 {
		recommendations = append(recommendations, "High request rate - monitor for overload")
	}

	// Selection recommendations
	if analytics.SelectionCount == 0 {
		recommendations = append(recommendations, "Carrier never selected - investigate configuration")
	} else if analytics.AverageScore < 70 {
		recommendations = append(recommendations, "Low selection scores - review carrier performance")
	}

	// Health recommendations
	if analytics.Status != "healthy" {
		recommendations = append(recommendations, "Carrier health issues - immediate attention required")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Carrier performing well")
	}

	return recommendations
}

// OptimizeCarrierSelection provides optimization recommendations
func (s *SelectionService) OptimizeCarrierSelection(ctx context.Context) (*OptimizationResponse, error) {
	s.logger.Info("Analyzing carrier selection optimization opportunities")

	analytics, err := s.GetCarrierAnalytics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	optimization := &OptimizationResponse{
		GeneratedAt:     time.Now(),
		OverallHealth:   analytics.Summary.OverallHealth,
		Recommendations: []string{},
		PriorityActions: []string{},
		LongTermActions: []string{},
	}

	// Generate recommendations based on analytics
	if analytics.Summary.OverallHealth < 80 {
		optimization.Recommendations = append(optimization.Recommendations, "System health below optimal - investigate carrier issues")
		optimization.PriorityActions = append(optimization.PriorityActions, "Review unhealthy carriers and implement failover")
	}

	if analytics.Summary.OverallSuccessRate < 95 {
		optimization.Recommendations = append(optimization.Recommendations, "Success rate below target - optimize carrier selection criteria")
		optimization.PriorityActions = append(optimization.PriorityActions, "Adjust performance weights in selection algorithm")
	}

	// Analyze individual carriers
	for _, carrier := range analytics.Analytics {
		if carrier.SuccessRate < 90 {
			optimization.PriorityActions = append(optimization.PriorityActions,
				fmt.Sprintf("Investigate %s (%s) - success rate %.1f%%",
					carrier.CarrierName, carrier.CarrierID, carrier.SuccessRate))
		}
		if carrier.AverageResponseTime > 1000 {
			optimization.PriorityActions = append(optimization.PriorityActions,
				fmt.Sprintf("Optimize %s (%s) - response time %dms",
					carrier.CarrierName, carrier.CarrierID, carrier.AverageResponseTime))
		}
	}

	// Long-term recommendations
	optimization.LongTermActions = append(optimization.LongTermActions,
		"Implement machine learning for carrier selection")
	optimization.LongTermActions = append(optimization.LongTermActions,
		"Add predictive analytics for carrier performance")
	optimization.LongTermActions = append(optimization.LongTermActions,
		"Expand carrier portfolio for better redundancy")

	if len(optimization.Recommendations) == 0 {
		optimization.Recommendations = append(optimization.Recommendations, "System operating optimally")
	}

	return optimization, nil
}

// OptimizationResponse represents optimization recommendations
type OptimizationResponse struct {
	GeneratedAt     time.Time `json:"generated_at"`
	OverallHealth   float64   `json:"overall_health"`
	Recommendations []string  `json:"recommendations"`
	PriorityActions []string  `json:"priority_actions"`
	LongTermActions []string  `json:"long_term_actions"`
}

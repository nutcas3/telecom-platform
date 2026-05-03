package handlers

import (
	"net/http"
	"time"
)

// GetSelectionAnalytics handles the selection analytics endpoint
func (h *SelectionHandler) GetSelectionAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all carriers
	carriers := h.manager.GetCarrierStatus()

	analytics := make(map[string]any)
	analytics["generated_at"] = time.Now().Format(time.RFC3339)
	analytics["total_carriers"] = len(carriers)

	// Calculate health distribution
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, carrier := range carriers {
		switch carrier.HealthStatus {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "unhealthy":
			unhealthyCount++
		}
	}

	healthDistribution := map[string]int{
		"healthy":   healthyCount,
		"degraded":  degradedCount,
		"unhealthy": unhealthyCount,
	}
	analytics["health_distribution"] = healthDistribution

	// Get learning statistics (using selection algorithm)
	learningStats := map[string]any{
		"learning_enabled": true,
		"message":          "Learning statistics available through selection algorithm",
	}
	analytics["learning_stats"] = learningStats

	// Calculate overall health percentage
	if len(carriers) > 0 {
		overallHealth := float64(healthyCount) / float64(len(carriers)) * 100
		analytics["overall_health_percentage"] = overallHealth
	}

	response := map[string]any{
		"success":   true,
		"analytics": analytics,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

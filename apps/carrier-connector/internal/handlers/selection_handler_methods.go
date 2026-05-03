package handlers

import (
	"encoding/json"
	"net/http"
)

// GetSelectionHistory handles the selection history endpoint
func (h *SelectionHandler) GetSelectionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract carrier ID from URL path
	carrierID := h.extractCarrierIDFromPath(r.URL.Path, "/api/v1/selection/history/")

	if carrierID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Carrier ID is required")
		return
	}

	history := h.manager.GetSelectionHistory(carrierID)

	// Generate analytics
	analytics := h.generateHistoryAnalytics(history)

	response := SelectionHistoryResponse{
		Success:   true,
		History:   history,
		Count:     len(history),
		Analytics: analytics,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateLearning handles the learning update endpoint
func (h *SelectionHandler) UpdateLearning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req UpdateLearningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.CarrierID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Carrier ID is required")
		return
	}
	if req.ActualPerformance < 0 || req.ActualPerformance > 100 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Actual performance must be between 0 and 100")
		return
	}

	// Update learning
	h.manager.UpdateLearning(req.CarrierID, req.ActualPerformance)

	response := UpdateLearningResponse{
		Success: true,
		Message: "Learning updated successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

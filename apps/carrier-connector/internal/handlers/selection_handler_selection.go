package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// SelectOptimalCarrier handles the optimal carrier selection endpoint
func (h *SelectionHandler) SelectOptimalCarrier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SelectOptimalCarrierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Urgency == "" {
		req.Urgency = "medium"
	}
	if req.CostSensitivity == 0 {
		req.CostSensitivity = 0.5
	}
	if req.PerformanceWeight == 0 {
		req.PerformanceWeight = 0.4
	}
	if req.ReliabilityWeight == 0 {
		req.ReliabilityWeight = 0.4
	}

	criteria := &smdp.SelectionCriteria{
		Region:            req.Region,
		ProfileType:       req.ProfileType,
		Urgency:           req.Urgency,
		CostSensitivity:   req.CostSensitivity,
		PerformanceWeight: req.PerformanceWeight,
		ReliabilityWeight: req.ReliabilityWeight,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	score, err := h.manager.SelectOptimalCarrier(ctx, criteria)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate recommendations
	recommendations := h.generateRecommendations(score, criteria)

	response := SelectOptimalCarrierResponse{
		Success:          true,
		CarrierID:        score.CarrierID,
		CarrierName:      score.Carrier.Name,
		TotalScore:       score.TotalScore,
		PerformanceScore: score.PerformanceScore,
		ReliabilityScore: score.ReliabilityScore,
		CostScore:        score.CostScore,
		RegionScore:      score.RegionScore,
		CapabilityScore:  score.CapabilityScore,
		SelectedAt:       score.SelectedAt.Format(time.RFC3339),
		Reason:           score.Reason,
		Recommendations:  recommendations,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// SelectCarrier handles the default carrier selection endpoint
func (h *SelectionHandler) SelectCarrier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	carrier, err := h.manager.SelectCarrier(ctx)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := SelectCarrierResponse{
		Success:     true,
		CarrierID:   carrier.ID,
		CarrierName: carrier.Name,
		Message:     "Default carrier selected successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

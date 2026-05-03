package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
)

// SelectionHandler handles carrier selection API endpoints
type SelectionHandler struct {
	manager *smdp.SMDPManager
	logger  *logrus.Logger
}

// NewSelectionHandler creates a new selection handler
func NewSelectionHandler(manager *smdp.SMDPManager) *SelectionHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &SelectionHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers selection-related routes
func (h *SelectionHandler) RegisterRoutes(mux *http.ServeMux) {
	// Carrier selection endpoints
	mux.HandleFunc("/api/v1/selection/optimal", h.SelectOptimalCarrier)
	mux.HandleFunc("/api/v1/selection/carrier", h.SelectCarrier)
	mux.HandleFunc("/api/v1/selection/history/", h.GetSelectionHistory)
	mux.HandleFunc("/api/v1/selection/learning", h.UpdateLearning)

	// Analytics endpoints
	mux.HandleFunc("/api/v1/selection/analytics/selection", h.GetSelectionAnalytics)
	mux.HandleFunc("/api/v1/selection/analytics/performance", h.GetPerformanceAnalytics)
}

// SelectOptimalCarrierRequest represents the request for optimal carrier selection
type SelectOptimalCarrierRequest struct {
	Region            string  `json:"region"`
	ProfileType       string  `json:"profile_type"`
	Urgency           string  `json:"urgency"`
	CostSensitivity   float64 `json:"cost_sensitivity"`
	PerformanceWeight float64 `json:"performance_weight"`
	ReliabilityWeight float64 `json:"reliability_weight"`
}

// SelectOptimalCarrierResponse represents the response for optimal carrier selection
type SelectOptimalCarrierResponse struct {
	Success          bool           `json:"success"`
	CarrierID        string         `json:"carrier_id"`
	CarrierName      string         `json:"carrier_name"`
	TotalScore       float64        `json:"total_score"`
	PerformanceScore float64        `json:"performance_score"`
	ReliabilityScore float64        `json:"reliability_score"`
	CostScore        float64        `json:"cost_score"`
	RegionScore      float64        `json:"region_score"`
	CapabilityScore  float64        `json:"capability_score"`
	SelectedAt       string         `json:"selected_at"`
	Reason           string         `json:"reason"`
	Recommendations  []string       `json:"recommendations"`
	Analytics        map[string]any `json:"analytics,omitempty"`
}

// SelectCarrierResponse represents the response for default carrier selection
type SelectCarrierResponse struct {
	Success     bool   `json:"success"`
	CarrierID   string `json:"carrier_id"`
	CarrierName string `json:"carrier_name"`
	Message     string `json:"message"`
}

// SelectionHistoryResponse represents the response for selection history
type SelectionHistoryResponse struct {
	Success   bool                `json:"success"`
	History   []smdp.CarrierScore `json:"history"`
	Count     int                 `json:"count"`
	Analytics map[string]any      `json:"analytics,omitempty"`
}

// UpdateLearningRequest represents the request for updating learning
type UpdateLearningRequest struct {
	CarrierID         string  `json:"carrier_id"`
	ActualPerformance float64 `json:"actual_performance"`
}

// UpdateLearningResponse represents the response for learning updates
type UpdateLearningResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Stats   map[string]any `json:"stats,omitempty"`
}

// writeJSONResponse writes a JSON response
func (h *SelectionHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes an error response
func (h *SelectionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{
		Success: false,
		Error:   message,
	}

	h.writeJSONResponse(w, statusCode, response)
}

// extractCarrierIDFromPath extracts carrier ID from URL path
func (h *SelectionHandler) extractCarrierIDFromPath(path, prefix string) string {
	carrierID := strings.TrimPrefix(path, prefix)
	carrierID = strings.TrimSuffix(carrierID, "/")
	return carrierID
}

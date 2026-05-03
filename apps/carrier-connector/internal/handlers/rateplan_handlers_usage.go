package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// RecordUsageRequest represents the request to record usage
type RecordUsageRequest struct {
	ProfileID string `json:"profile_id" binding:"required"`
	DataUsed  int64  `json:"data_used" binding:"required,min=0"`
	VoiceUsed int64  `json:"voice_used" binding:"required,min=0"`
	SMSUsed   int64  `json:"sms_used" binding:"required,min=0"`
}

// UsageResponse represents the response for usage operations
type UsageResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Data    *rateplan.RatePlanUsage `json:"data,omitempty"`
}

// UsageHistoryResponse represents the response for usage history
type UsageHistoryResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message,omitempty"`
	Data    []*rateplan.RatePlanUsage `json:"data,omitempty"`
}

// RecordUsage handles recording usage for a subscription
func (h *RatePlanHandler) RecordUsage(c *gin.Context) {
	var req RecordUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	recordReq := &rateplan.RecordUsageRequest{
		ProfileID: req.ProfileID,
		DataUsed:  req.DataUsed,
		VoiceUsed: req.VoiceUsed,
		SMSUsed:   req.SMSUsed,
	}

	usage, err := h.service.RecordUsage(c.Request.Context(), recordReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to record usage")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to record usage: "+err.Error())
		return
	}

	response := UsageResponse{
		Success: true,
		Message: "Usage recorded successfully",
		Data:    usage,
	}

	h.writeJSONResponse(c, http.StatusCreated, response)
}

// GetUsage handles retrieving current usage for a profile
func (h *RatePlanHandler) GetUsage(c *gin.Context) {
	profileID := c.Query("profile_id")
	if profileID == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Profile ID is required")
		return
	}

	usage, err := h.service.GetUsage(c.Request.Context(), profileID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get usage")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get usage")
		return
	}

	if usage == nil {
		response := struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}{
			Success: false,
			Message: "No usage data found",
		}
		h.writeJSONResponse(c, http.StatusOK, response)
		return
	}

	response := UsageResponse{
		Success: true,
		Data:    usage,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// GetUsageHistory handles retrieving usage history for a profile
func (h *RatePlanHandler) GetUsageHistory(c *gin.Context) {
	profileID := c.Query("profile_id")
	if profileID == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Profile ID is required")
		return
	}

	limit := 10 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	usageHistory, err := h.service.GetUsageHistory(c.Request.Context(), profileID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get usage history")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get usage history")
		return
	}

	response := UsageHistoryResponse{
		Success: true,
		Data:    usageHistory,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// CalculateCostRequest represents the request to calculate cost
type CalculateCostRequest struct {
	RatePlanID       string   `json:"rate_plan_id" binding:"required"`
	DataUsed         int64    `json:"data_used" binding:"required,min=0"`
	VoiceUsed        int64    `json:"voice_used" binding:"required,min=0"`
	SMSUsed          int64    `json:"sms_used" binding:"required,min=0"`
	AppliedDiscounts []string `json:"applied_discounts,omitempty"`
}

// CostCalculationResponse represents the response for cost calculation
type CostCalculationResponse struct {
	Success bool                         `json:"success"`
	Message string                       `json:"message,omitempty"`
	Data    *rateplan.RatePlanCostCalculation `json:"data,omitempty"`
}

// CalculateCost handles calculating cost for a rate plan based on usage
func (h *RatePlanHandler) CalculateCost(c *gin.Context) {
	var req CalculateCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	calcReq := &rateplan.CalculateCostRequest{
		RatePlanID:       req.RatePlanID,
		DataUsed:         req.DataUsed,
		VoiceUsed:        req.VoiceUsed,
		SMSUsed:          req.SMSUsed,
		AppliedDiscounts: req.AppliedDiscounts,
	}

	calculation, err := h.service.CalculateCost(c.Request.Context(), calcReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate cost")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to calculate cost: "+err.Error())
		return
	}

	response := CostCalculationResponse{
		Success: true,
		Data:    calculation,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

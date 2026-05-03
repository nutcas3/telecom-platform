package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// CreateRatePlan handles the creation of a new rate plan
func (h *RatePlanHandler) CreateRatePlan(c *gin.Context) {
	var req CreateRatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Convert request to rate plan
	plan := &rateplan.RatePlan{
		Name:              req.Name,
		Description:       req.Description,
		CarrierID:         req.CarrierID,
		Region:            req.Region,
		PlanType:          req.PlanType,
		Status:            rateplan.PlanStatusDraft,
		BasePrice:         req.BasePrice,
		Currency:          req.Currency,
		BillingCycle:      req.BillingCycle,
		DataAllowance:     req.DataAllowance,
		VoiceAllowance:    req.VoiceAllowance,
		SMSAllowance:      req.SMSAllowance,
		OverageRates:      req.OverageRates,
		Features:          req.Features,
		ActivationFee:     req.ActivationFee,
		EarlyTermination:  req.EarlyTermination,
		Discounts:         req.Discounts,
		ValidFrom:         req.ValidFrom,
		ValidTo:           req.ValidTo,
		Priority:          req.Priority,
		IsActive:          req.IsActive,
		Metadata:          req.Metadata,
	}

	// Create the rate plan
	createdPlan, err := h.service.CreateRatePlan(c.Request.Context(), plan)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create rate plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to create rate plan: "+err.Error())
		return
	}

	response := RatePlanResponse{
		Success: true,
		Message: "Rate plan created successfully",
		Data:    createdPlan,
	}

	h.writeJSONResponse(c, http.StatusCreated, response)
}

// GetRatePlan handles retrieving a rate plan by ID
func (h *RatePlanHandler) GetRatePlan(c *gin.Context) {
	id := h.extractIDFromPath(c)
	if id == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Rate plan ID is required")
		return
	}

	plan, err := h.service.GetRatePlan(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "rate plan not found: "+id {
			h.writeErrorResponse(c, http.StatusNotFound, "Rate plan not found")
			return
		}
		h.logger.WithError(err).Error("Failed to get rate plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get rate plan")
		return
	}

	response := RatePlanResponse{
		Success: true,
		Data:    plan,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// UpdateRatePlan handles updating an existing rate plan
func (h *RatePlanHandler) UpdateRatePlan(c *gin.Context) {
	id := h.extractIDFromPath(c)
	if id == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Rate plan ID is required")
		return
	}

	var req UpdateRatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Get existing rate plan
	plan, err := h.service.GetRatePlan(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "rate plan not found: "+id {
			h.writeErrorResponse(c, http.StatusNotFound, "Rate plan not found")
			return
		}
		h.logger.WithError(err).Error("Failed to get rate plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get rate plan")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.PlanType != "" {
		plan.PlanType = req.PlanType
	}
	if req.Status != "" {
		plan.Status = req.Status
	}
	if req.BasePrice > 0 {
		plan.BasePrice = req.BasePrice
	}
	if req.Currency != "" {
		plan.Currency = req.Currency
	}
	if req.BillingCycle != "" {
		plan.BillingCycle = req.BillingCycle
	}
	if req.DataAllowance != nil {
		plan.DataAllowance = req.DataAllowance
	}
	if req.VoiceAllowance != nil {
		plan.VoiceAllowance = req.VoiceAllowance
	}
	if req.SMSAllowance != nil {
		plan.SMSAllowance = req.SMSAllowance
	}
	if req.OverageRates != nil {
		plan.OverageRates = req.OverageRates
	}
	if req.Features != nil {
		plan.Features = req.Features
	}
	if req.ActivationFee >= 0 {
		plan.ActivationFee = req.ActivationFee
	}
	if req.EarlyTermination != nil {
		plan.EarlyTermination = req.EarlyTermination
	}
	if req.Discounts != nil {
		plan.Discounts = req.Discounts
	}
	if req.ValidFrom != nil {
		plan.ValidFrom = *req.ValidFrom
	}
	if req.ValidTo != nil {
		plan.ValidTo = req.ValidTo
	}
	if req.Priority != 0 {
		plan.Priority = req.Priority
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		plan.Metadata = req.Metadata
	}

	// Update the rate plan
	updatedPlan, err := h.service.UpdateRatePlan(c.Request.Context(), plan)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update rate plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to update rate plan: "+err.Error())
		return
	}

	response := RatePlanResponse{
		Success: true,
		Message: "Rate plan updated successfully",
		Data:    updatedPlan,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// DeleteRatePlan handles deleting a rate plan
func (h *RatePlanHandler) DeleteRatePlan(c *gin.Context) {
	id := h.extractIDFromPath(c)
	if id == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Rate plan ID is required")
		return
	}

	if err := h.service.DeleteRatePlan(c.Request.Context(), id); err != nil {
		if err.Error() == "rate plan not found: "+id {
			h.writeErrorResponse(c, http.StatusNotFound, "Rate plan not found")
			return
		}
		if err.Error() == "cannot delete rate plan with active subscriptions" {
			h.writeErrorResponse(c, http.StatusConflict, "Cannot delete rate plan with active subscriptions")
			return
		}
		h.logger.WithError(err).Error("Failed to delete rate plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to delete rate plan")
		return
	}

	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Rate plan deleted successfully",
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// ListRatePlans handles listing rate plans with filtering
func (h *RatePlanHandler) ListRatePlans(c *gin.Context) {
	// Parse query parameters
	filter := &rateplan.RatePlanFilter{
		CarrierID: c.Query("carrier_id"),
		Region:    c.Query("region"),
		PlanType:  rateplan.PlanType(c.Query("plan_type")),
		Status:    rateplan.PlanStatus(c.Query("status")),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Parse numeric parameters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = maxPrice
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Parse boolean parameter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Get rate plans
	plans, err := h.service.ListRatePlans(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list rate plans")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to list rate plans")
		return
	}

	response := RatePlansResponse{
		Success: true,
		Data:    plans,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// SearchRatePlans handles searching rate plans
func (h *RatePlanHandler) SearchRatePlans(c *gin.Context) {
	criteria := rateplan.SearchCriteria{
		CarrierID: c.Query("carrier_id"),
		Region:    c.Query("region"),
		PlanType:  rateplan.PlanType(c.Query("plan_type")),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Parse numeric parameters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			criteria.MinPrice = minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			criteria.MaxPrice = maxPrice
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			criteria.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			criteria.Offset = offset
		}
	}

	plans, err := h.service.SearchRatePlans(c.Request.Context(), criteria)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search rate plans")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to search rate plans")
		return
	}

	response := RatePlansResponse{
		Success: true,
		Data:    plans,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

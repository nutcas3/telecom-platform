package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// SubscribeRequest represents the request to subscribe to a rate plan
type SubscribeRequest struct {
	ProfileID        string         `json:"profile_id" binding:"required"`
	RatePlanID       string         `json:"rate_plan_id" binding:"required"`
	AutoRenew        bool           `json:"auto_renew"`
	AppliedDiscounts []string       `json:"applied_discounts,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// SubscriptionResponse represents the response for subscription operations
type SubscriptionResponse struct {
	Success bool                           `json:"success"`
	Message string                         `json:"message,omitempty"`
	Data    *rateplan.RatePlanSubscription `json:"data,omitempty"`
}

// SubscriptionsResponse represents the response for listing subscriptions
type SubscriptionsResponse struct {
	Success bool                             `json:"success"`
	Message string                           `json:"message,omitempty"`
	Data    []*rateplan.RatePlanSubscription `json:"data,omitempty"`
}

// SubscribeToPlan handles subscribing a profile to a rate plan
func (h *RatePlanHandler) SubscribeToPlan(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	subscribeReq := &rateplan.SubscribeRequest{
		ProfileID:        req.ProfileID,
		RatePlanID:       req.RatePlanID,
		AutoRenew:        req.AutoRenew,
		AppliedDiscounts: req.AppliedDiscounts,
		Metadata:         req.Metadata,
	}

	subscription, err := h.service.SubscribeToPlan(c.Request.Context(), subscribeReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to subscribe to plan")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to subscribe to plan: "+err.Error())
		return
	}

	response := SubscriptionResponse{
		Success: true,
		Message: "Subscription created successfully",
		Data:    subscription,
	}

	h.writeJSONResponse(c, http.StatusCreated, response)
}

// GetSubscription handles retrieving a subscription by ID
func (h *RatePlanHandler) GetSubscription(c *gin.Context) {
	id := h.extractIDFromPath(c)
	if id == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	subscription, err := h.service.GetSubscription(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "subscription not found: "+id {
			h.writeErrorResponse(c, http.StatusNotFound, "Subscription not found")
			return
		}
		h.logger.WithError(err).Error("Failed to get subscription")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription")
		return
	}

	response := SubscriptionResponse{
		Success: true,
		Data:    subscription,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// GetActiveSubscription handles retrieving the active subscription for a profile
func (h *RatePlanHandler) GetActiveSubscription(c *gin.Context) {
	profileID := c.Query("profile_id")
	if profileID == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Profile ID is required")
		return
	}

	subscription, err := h.service.GetActiveSubscription(c.Request.Context(), profileID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get active subscription")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get active subscription")
		return
	}

	if subscription == nil {
		response := struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}{
			Success: false,
			Message: "No active subscription found",
		}
		h.writeJSONResponse(c, http.StatusOK, response)
		return
	}

	response := SubscriptionResponse{
		Success: true,
		Data:    subscription,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// ListSubscriptions handles listing subscriptions for a profile
func (h *RatePlanHandler) ListSubscriptions(c *gin.Context) {
	profileID := c.Query("profile_id")
	if profileID == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Profile ID is required")
		return
	}

	filter := &rateplan.SubscriptionFilter{
		Status:     rateplan.SubscriptionStatus(c.Query("status")),
		RatePlanID: c.Query("rate_plan_id"),
	}

	// Parse limit and offset
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

	subscriptions, err := h.service.ListSubscriptions(c.Request.Context(), profileID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to list subscriptions")
		return
	}

	response := SubscriptionsResponse{
		Success: true,
		Data:    subscriptions,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// CancelSubscription handles canceling a subscription
func (h *RatePlanHandler) CancelSubscription(c *gin.Context) {
	id := h.extractIDFromPath(c)
	if id == "" {
		h.writeErrorResponse(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.service.CancelSubscription(c.Request.Context(), id, req.Reason); err != nil {
		if err.Error() == "subscription not found: "+id {
			h.writeErrorResponse(c, http.StatusNotFound, "Subscription not found")
			return
		}
		if err.Error() == "subscription is not active" {
			h.writeErrorResponse(c, http.StatusBadRequest, "Subscription is not active")
			return
		}
		h.logger.WithError(err).Error("Failed to cancel subscription")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to cancel subscription")
		return
	}

	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Subscription cancelled successfully",
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

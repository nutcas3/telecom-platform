package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// AnalyticsResponse represents the response for analytics operations
type AnalyticsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// GetUsageAnalytics handles retrieving usage analytics
func (h *RatePlanHandler) GetUsageAnalytics(c *gin.Context) {
	// Parse query parameters
	filter := &rateplan.UsageAnalyticsFilter{
		RatePlanID: c.Query("rate_plan_id"),
		CarrierID:  c.Query("carrier_id"),
		Region:     c.Query("region"),
		GroupBy:    c.Query("group_by"),
	}

	// Parse date parameters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = endDate
		}
	}

	// Set default date range if not provided
	if filter.StartDate.IsZero() {
		filter.StartDate = time.Now().AddDate(0, -1, 0) // Last month
	}
	if filter.EndDate.IsZero() {
		filter.EndDate = time.Now()
	}

	analytics, err := h.service.GetUsageAnalytics(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get usage analytics")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get usage analytics")
		return
	}

	response := AnalyticsResponse{
		Success: true,
		Data:    analytics,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// GetRevenueAnalytics handles retrieving revenue analytics
func (h *RatePlanHandler) GetRevenueAnalytics(c *gin.Context) {
	// Parse query parameters
	filter := &rateplan.RevenueAnalyticsFilter{
		RatePlanID: c.Query("rate_plan_id"),
		CarrierID:  c.Query("carrier_id"),
		Region:     c.Query("region"),
		GroupBy:    c.Query("group_by"),
	}

	// Parse date parameters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = endDate
		}
	}

	// Set default date range if not provided
	if filter.StartDate.IsZero() {
		filter.StartDate = time.Now().AddDate(0, -1, 0) // Last month
	}
	if filter.EndDate.IsZero() {
		filter.EndDate = time.Now()
	}

	analytics, err := h.service.GetRevenueAnalytics(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get revenue analytics")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get revenue analytics")
		return
	}

	response := AnalyticsResponse{
		Success: true,
		Data:    analytics,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// GetPopularPlans handles retrieving the most popular rate plans
func (h *RatePlanHandler) GetPopularPlans(c *gin.Context) {
	limit := 10 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	plans, err := h.service.GetPopularPlans(c.Request.Context(), limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get popular plans")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get popular plans")
		return
	}

	response := RatePlansResponse{
		Success: true,
		Data:    plans,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

// GetDashboardData handles retrieving dashboard data for rate plans
func (h *RatePlanHandler) GetDashboardData(c *gin.Context) {
	// Get popular plans (top 5)
	popularPlans, err := h.service.GetPopularPlans(c.Request.Context(), 5)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get popular plans for dashboard")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard data")
		return
	}

	// Get usage analytics for last 30 days
	usageFilter := &rateplan.UsageAnalyticsFilter{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	usageAnalytics, err := h.service.GetUsageAnalytics(c.Request.Context(), usageFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get usage analytics for dashboard")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard data")
		return
	}

	// Get revenue analytics for last 30 days
	revenueFilter := &rateplan.RevenueAnalyticsFilter{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	revenueAnalytics, err := h.service.GetRevenueAnalytics(c.Request.Context(), revenueFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get revenue analytics for dashboard")
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard data")
		return
	}

	dashboardData := map[string]any{
		"popular_plans":     popularPlans,
		"usage_analytics":   usageAnalytics,
		"revenue_analytics": revenueAnalytics,
		"generated_at":      time.Now().Format(time.RFC3339),
	}

	response := AnalyticsResponse{
		Success: true,
		Data:    dashboardData,
	}

	h.writeJSONResponse(c, http.StatusOK, response)
}

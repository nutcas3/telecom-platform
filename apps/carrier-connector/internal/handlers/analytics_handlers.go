package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/analytics"
	"github.com/sirupsen/logrus"
)

// AnalyticsHandler handles analytics API requests
type AnalyticsHandler struct {
	service *analytics.Service
	logger  *logrus.Logger
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service *analytics.Service, logger *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

// GetDashboard returns the main analytics dashboard
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}

	filter := &analytics.AnalyticsFilter{
		TenantID:  tenantID,
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
	}

	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			filter.StartDate = t
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			filter.EndDate = t
		}
	}

	dashboard, err := h.service.GetDashboard(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetRevenueAnalytics returns detailed revenue analytics
func (h *AnalyticsHandler) GetRevenueAnalytics(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	filter := &analytics.AnalyticsFilter{
		TenantID:  tenantID,
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
		GroupBy:   c.Query("group_by"),
	}

	revenue, err := h.service.GetRevenueAnalytics(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, revenue)
}

// CreateScheduledReport creates a scheduled report
func (h *AnalyticsHandler) CreateScheduledReport(c *gin.Context) {
	var report analytics.ScheduledReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.CreateScheduledReport(c.Request.Context(), &report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// ListScheduledReports lists scheduled reports
func (h *AnalyticsHandler) ListScheduledReports(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	reports, err := h.service.ListScheduledReports(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

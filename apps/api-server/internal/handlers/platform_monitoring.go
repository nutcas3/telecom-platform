package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/monitoring"
)

// MonitoringHandler exposes Prometheus-backed metrics and alert endpoints.
type MonitoringHandler struct {
	prom *monitoring.PrometheusService
}

// NewMonitoringHandler constructs a MonitoringHandler.
func NewMonitoringHandler(prom *monitoring.PrometheusService) *MonitoringHandler {
	return &MonitoringHandler{prom: prom}
}

func (h *MonitoringHandler) Metrics(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitoring service not available"})
		return
	}
	q := c.DefaultQuery("query", "up")
	samples, err := h.prom.Query(c.Request.Context(), q)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"query": q, "samples": samples})
}

func (h *MonitoringHandler) Alerts(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitoring service not available"})
		return
	}
	alerts, err := h.prom.Alerts(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (h *MonitoringHandler) Health(c *gin.Context) {
	if h.prom == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":               "degraded",
			"prometheus_available": false,
			"error":                "monitoring service not available",
		})
		return
	}
	available := h.prom.Available(c.Request.Context())
	status := "healthy"
	if !available {
		status = "degraded"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":               status,
		"prometheus_available": available,
	})
}

func (h *MonitoringHandler) Logs(c *gin.Context) {
	// Logs are aggregated in Elasticsearch; expose a stub that points callers
	// to /v1/services/:id/logs for live pod logs.
	c.JSON(http.StatusOK, gin.H{
		"message": "Use /v1/services/:id/logs for live pod logs; aggregated logs are in Kibana",
	})
}

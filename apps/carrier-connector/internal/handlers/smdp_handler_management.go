package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// AddCarrier handles adding a new carrier
func (h *SMDPHandler) AddCarrier(c *gin.Context) {
	var carrierConfig smdp.CarrierConfig
	if err := c.ShouldBindJSON(&carrierConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate carrier config
	if carrierConfig.ID == "" || carrierConfig.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Carrier ID and name are required",
		})
		return
	}

	carrier := carrierConfig.ToCarrier()
	if err := h.manager.AddCarrier(carrier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add carrier",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Carrier added successfully",
		"carrier": carrier,
	})
}

// GetAnalytics handles analytics requests
func (h *SMDPHandler) GetAnalytics(c *gin.Context) {
	// Get basic analytics
	carriers := h.manager.GetCarrierStatus()
	totalCarriers := len(carriers)
	healthyCarriers := 0

	for _, carrier := range carriers {
		if carrier.HealthStatus == "healthy" {
			healthyCarriers++
		}
	}

	overallHealth := 0.0
	if totalCarriers > 0 {
		overallHealth = float64(healthyCarriers) / float64(totalCarriers) * 100
	}

	analytics := map[string]any{
		"generated_at":      time.Now().Format(time.RFC3339),
		"total_carriers":    totalCarriers,
		"healthy_carriers":  healthyCarriers,
		"overall_health":    overallHealth,
		"health_percentage": overallHealth,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"analytics": analytics,
	})
}

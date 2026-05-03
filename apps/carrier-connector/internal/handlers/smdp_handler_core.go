package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// SMDPHandler handles SM-DP+ related operations
type SMDPHandler struct {
	manager *smdp.SMDPManager
}

// NewSMDPHandler creates a new SM-DP+ handler
func NewSMDPHandler(manager *smdp.SMDPManager) *SMDPHandler {
	return &SMDPHandler{
		manager: manager,
	}
}

// DownloadProfile handles profile download requests
func (h *SMDPHandler) DownloadProfile(c *gin.Context) {
	var req smdp.ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.EID == "" && req.ICCID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "EID or ICCID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.manager.DownloadProfile(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to download profile",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       response.Success,
		"response_time": response.ResponseTime.Milliseconds(),
		"message":       response.StatusMessage,
	})
}

// GetCarrierStatus handles carrier status requests
func (h *SMDPHandler) GetCarrierStatus(c *gin.Context) {
	carriers := h.manager.GetCarrierStatus()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"carriers": carriers,
		"total":    len(carriers),
	})
}

// GetCarrierHealth handles carrier health check requests
func (h *SMDPHandler) GetCarrierHealth(c *gin.Context) {
	carrierID := c.Param("carrier_id")
	if carrierID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Carrier ID is required"})
		return
	}

	carriers := h.manager.GetCarrierStatus()
	carrier, exists := carriers[carrierID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Carrier not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"carrier": carrier,
	})
}

// GetProfileStatus handles profile status check requests
func (h *SMDPHandler) GetProfileStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile status endpoint",
		"status":  "active",
	})
}

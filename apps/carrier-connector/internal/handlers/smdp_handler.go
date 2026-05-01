package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// SMDPHandler handles SM-DP+ multi-carrier operations
type SMDPHandler struct {
	manager *smdp.SMDPManager
}

// NewSMDPHandler creates a new SM-DP+ handler
func NewSMDPHandler(repo *repository.PostgresProfileStore) *SMDPHandler {
	config := smdp.DefaultManagerConfig()
	manager := smdp.NewSMDPManager(repo, config)

	// Start health checking in background
	ctx := context.Background()
	go manager.StartHealthChecking(ctx)

	return &SMDPHandler{
		manager: manager,
	}
}

// AddCarrier adds a new carrier to the SM-DP+ manager
func (h *SMDPHandler) AddCarrier(c *gin.Context) {
	var carrierConfig smdp.CarrierConfig
	if err := c.ShouldBindJSON(&carrierConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	carrier := carrierConfig.ToCarrier()
	if err := h.manager.AddCarrier(carrier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Carrier added successfully",
		"carrier_id": carrier.ID,
	})
}

// RemoveCarrier removes a carrier from the SM-DP+ manager
func (h *SMDPHandler) RemoveCarrier(c *gin.Context) {
	carrierID := c.Param("id")
	if carrierID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Carrier ID is required"})
		return
	}

	if err := h.manager.RemoveCarrier(carrierID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Carrier removed successfully"})
}

// DownloadProfile handles eSIM profile download with multi-carrier support
func (h *SMDPHandler) DownloadProfile(c *gin.Context) {
	var req smdp.ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.manager.DownloadProfile(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": response,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCarrierStatus returns the status of all carriers
func (h *SMDPHandler) GetCarrierStatus(c *gin.Context) {
	status := h.manager.GetCarrierStatus()
	c.JSON(http.StatusOK, gin.H{
		"carriers":  status,
		"timestamp": time.Now(),
	})
}

// GetProfileStatus gets profile status from the best available carrier
func (h *SMDPHandler) GetProfileStatus(c *gin.Context) {
	var req struct {
		EID   string `json:"eid" binding:"required"`
		ICCID string `json:"iccid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create profile request for status check
	profileReq := &smdp.ProfileRequest{
		EID:   req.EID,
		ICCID: req.ICCID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Use the manager to select best carrier and get status
	carrier, err := h.manager.SelectCarrier(ctx, profileReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"carrier_id":    carrier.ID,
		"carrier_name":  carrier.Name,
		"health_status": carrier.HealthStatus,
		"message":       "Use specific carrier endpoint for detailed status",
	})
}

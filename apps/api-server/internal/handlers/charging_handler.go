package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ChargingHandler handles charging engine integration
type ChargingHandler struct {
	engineClient *services.ChargingEngineClient
}

// NewChargingHandler creates a new charging handler
func NewChargingHandler(engineClient *services.ChargingEngineClient) *ChargingHandler {
	return &ChargingHandler{
		engineClient: engineClient,
	}
}

// CheckCredit checks if an IP has sufficient credit
func (h *ChargingHandler) CheckCredit(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	var req struct {
		BytesRequested uint64 `json:"bytes_requested" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.engineClient.CheckCredit(c.Request.Context(), ip, req.BytesRequested)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetBalance gets the current balance for an IP
func (h *ChargingHandler) GetBalance(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	result, err := h.engineClient.GetBalance(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AddCredit adds credit to an IP
func (h *ChargingHandler) AddCredit(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	var req struct {
		BytesToAdd uint64 `json:"bytes_to_add" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.engineClient.AddCredit(c.Request.Context(), ip, req.BytesToAdd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeductCredit deducts credit from an IP
func (h *ChargingHandler) DeductCredit(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is required"})
		return
	}

	var req struct {
		BytesUsed uint64 `json:"bytes_used" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.engineClient.DeductCredit(c.Request.Context(), ip, req.BytesUsed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RemoveCarrier handles removing a carrier
func (h *SMDPHandler) RemoveCarrier(c *gin.Context) {
	carrierID := c.Param("carrier_id")
	if carrierID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Carrier ID is required"})
		return
	}

	if err := h.manager.RemoveCarrier(carrierID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to remove carrier",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Carrier removed successfully",
	})
}

// GetSelectionHistory handles selection history requests
func (h *SMDPHandler) GetSelectionHistory(c *gin.Context) {
	carrierID := c.Param("carrier_id")
	if carrierID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Carrier ID is required"})
		return
	}

	history := h.manager.GetSelectionHistory(carrierID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
		"count":   len(history),
	})
}

// UpdateLearning handles learning update requests
func (h *SMDPHandler) UpdateLearning(c *gin.Context) {
	carrierID := c.Param("carrier_id")
	if carrierID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Carrier ID is required"})
		return
	}

	var req struct {
		ActualPerformance float64 `json:"actual_performance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.ActualPerformance < 0 || req.ActualPerformance > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Actual performance must be between 0 and 100",
		})
		return
	}

	h.manager.UpdateLearning(carrierID, req.ActualPerformance)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Learning updated successfully",
	})
}

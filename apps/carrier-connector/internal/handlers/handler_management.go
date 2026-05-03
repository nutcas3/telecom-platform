package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mvno"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
)

// ManagementHandler handles MVNO management HTTP requests
type ManagementHandler struct {
	repo   repository.Repository
	logger *logrus.Logger
}

// NewManagementHandler creates a new management handler
func NewManagementHandler(repo repository.Repository, logger *logrus.Logger) *ManagementHandler {
	return &ManagementHandler{
		repo:   repo,
		logger: logger,
	}
}

// GetMVNO handles GET /mvno/{id}
func (h *ManagementHandler) GetMVNO(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MVNO ID is required"})
		return
	}

	mvno, err := h.repo.GetMVNO(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("mvno_id", id).Error("Failed to get MVNO")
		c.JSON(http.StatusNotFound, gin.H{"error": "MVNO not found"})
		return
	}

	c.JSON(http.StatusOK, mvno)
}

// ListMVNOs handles GET /mvno
func (h *ManagementHandler) ListMVNOs(c *gin.Context) {
	filter := &mvno.MVNOFilter{}

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		filter.Status = mvno.MVNOStatus(status)
	}
	if plan := c.Query("plan"); plan != "" {
		filter.Plan = mvno.MVNOPlan(plan)
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	mvnos, err := h.repo.ListMVNOs(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list MVNOs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list MVNOs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mvnos": mvnos,
		"count": len(mvnos),
	})
}

// UpdateMVNOStatus handles PUT /mvno/{id}/status
func (h *ManagementHandler) UpdateMVNOStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MVNO ID is required"})
		return
	}

	var req struct {
		Status mvno.MVNOStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateMVNOStatus(c.Request.Context(), id, req.Status); err != nil {
		h.logger.WithError(err).WithField("mvno_id", id).Error("Failed to update MVNO status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"mvno_id":    id,
		"new_status": req.Status,
	}).Info("MVNO status updated")

	c.JSON(http.StatusOK, gin.H{
		"mvno_id": id,
		"status":  req.Status,
	})
}

// GetMVNOStats handles GET /mvno/stats
func (h *ManagementHandler) GetMVNOStats(c *gin.Context) {
	stats, err := h.repo.GetMVNOStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get MVNO stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

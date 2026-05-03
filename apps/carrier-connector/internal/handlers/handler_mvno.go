package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mvno"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/sirupsen/logrus"
)

// OnboardingService defines the interface for MVNO onboarding operations
type OnboardingService interface {
	StartOnboarding(ctx context.Context, req *mvno.OnboardingRequest) (*mvno.MVNO, error)
}

// MVNOHandler handles HTTP requests for MVNO operations
type MVNOHandler struct {
	service OnboardingService
	repo    repository.Repository
	logger  *logrus.Logger
}

// NewMVNOHandler creates a new MVNO HTTP handler
func NewMVNOHandler(service OnboardingService, repo repository.Repository, logger *logrus.Logger) *MVNOHandler {
	return &MVNOHandler{
		service: service,
		repo:    repo,
		logger:  logger,
	}
}

// StartOnboarding handles POST /mvno/onboarding
func (h *MVNOHandler) StartOnboarding(c *gin.Context) {
	var req mvno.OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid onboarding request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mvno, err := h.service.StartOnboarding(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start onboarding")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateMVNO(c.Request.Context(), mvno); err != nil {
		h.logger.WithError(err).Error("Failed to save MVNO")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MVNO"})
		return
	}

	h.logger.WithField("mvno_id", mvno.ID).Info("Onboarding started")
	c.JSON(http.StatusCreated, gin.H{
		"mvno_id":     mvno.ID,
		"business_id": mvno.BusinessID,
		"status":      mvno.Status,
		"plan":        mvno.Plan,
	})
}

// GetMVNO handles GET /mvno/{id}
func (h *MVNOHandler) GetMVNO(c *gin.Context) {
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
func (h *MVNOHandler) ListMVNOs(c *gin.Context) {
	filter := &mvno.MVNOFilter{}
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

	c.JSON(http.StatusOK, gin.H{"mvnos": mvnos, "count": len(mvnos)})
}

// RegisterRoutes registers all MVNO routes
func (h *MVNOHandler) RegisterRoutes(router *gin.RouterGroup) {
	mvno := router.Group("/mvno")
	{
		mvno.POST("/onboarding", h.StartOnboarding)
		mvno.GET("", h.ListMVNOs)
		mvno.GET("/:id", h.GetMVNO)
	}
}

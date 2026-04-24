package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// DeploymentsHandler exposes deployment lifecycle endpoints.
type DeploymentsHandler struct {
	svc *services.DeploymentService
}

// NewDeploymentsHandler constructs a DeploymentsHandler.
func NewDeploymentsHandler(svc *services.DeploymentService) *DeploymentsHandler {
	return &DeploymentsHandler{svc: svc}
}

func (h *DeploymentsHandler) Status(c *gin.Context) {
	f := services.DeploymentFilter{
		Service:     c.Query("service"),
		Environment: c.Query("environment"),
		Status:      "in_progress",
		Limit:       50,
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deployments": items, "total": total})
}

func (h *DeploymentsHandler) Start(c *gin.Context) {
	var in services.StartDeploymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	d, err := h.svc.Start(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DeploymentsHandler) Rollback(c *gin.Context) {
	var in services.RollbackInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	d, err := h.svc.Rollback(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DeploymentsHandler) History(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}
	f := services.DeploymentFilter{
		Service:     c.Query("service"),
		Environment: c.Query("environment"),
		Status:      c.Query("status"),
		Limit:       pageSize,
		Offset:      (page - 1) * pageSize,
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"deployments": items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
	})
}

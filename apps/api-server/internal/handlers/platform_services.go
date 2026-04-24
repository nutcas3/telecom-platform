package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ServicesHandler exposes Kubernetes service/deployment operations.
type ServicesHandler struct {
	k8s *services.KubernetesService
}

// NewServicesHandler constructs a ServicesHandler.
func NewServicesHandler(k8s *services.KubernetesService) *ServicesHandler {
	return &ServicesHandler{k8s: k8s}
}

func (h *ServicesHandler) List(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	deps, err := h.k8s.ListDeployments(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"services": deps, "total": len(deps), "namespace": h.k8s.Namespace()})
}

func (h *ServicesHandler) Get(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	d, err := h.k8s.GetDeployment(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *ServicesHandler) Restart(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	if err := h.k8s.RestartDeployment(c.Request.Context(), c.Param("id")); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restart triggered", "service": c.Param("id")})
}

func (h *ServicesHandler) Start(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), 1); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service started", "service": c.Param("id")})
}

func (h *ServicesHandler) Stop(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), 0); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service stopped", "service": c.Param("id")})
}

func (h *ServicesHandler) Scale(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	var req struct {
		Replicas int32 `json:"replicas" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.k8s.ScaleDeployment(c.Request.Context(), c.Param("id"), req.Replicas); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "scaled", "service": c.Param("id"), "replicas": req.Replicas})
}

func (h *ServicesHandler) Logs(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))
	logs, err := h.k8s.PodLogs(c.Request.Context(), c.Param("id"), lines)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"service": c.Param("id"), "logs": logs})
}

func (h *ServicesHandler) Health(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	d, err := h.k8s.GetDeployment(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFound(c, err)
		return
	}
	status := "healthy"
	if d.ReadyReplicas < d.Replicas {
		status = "degraded"
	}
	if d.ReadyReplicas == 0 {
		status = "unhealthy"
	}
	c.JSON(http.StatusOK, gin.H{
		"service":            c.Param("id"),
		"status":             status,
		"replicas_desired":   d.Replicas,
		"replicas_ready":     d.ReadyReplicas,
		"replicas_available": d.AvailableReplicas,
	})
}

func (h *ServicesHandler) PodStatus(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	status, err := h.k8s.GetPodStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *ServicesHandler) Events(c *gin.Context) {
	if h.k8s == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Kubernetes not configured", Code: "K8S_UNAVAILABLE"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	events, err := h.k8s.GetEvents(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

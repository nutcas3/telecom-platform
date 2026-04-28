package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ChaosHandler exposes chaos-engineering experiment endpoints.
type ChaosHandler struct {
	svc *services.ChaosService
}

// NewChaosHandler constructs a ChaosHandler.
func NewChaosHandler(svc *services.ChaosService) *ChaosHandler {
	return &ChaosHandler{svc: svc}
}

type runExperimentRequest struct {
	Type        services.ExperimentType `json:"type" binding:"required"`
	Target      string                  `json:"target" binding:"required"`
	DurationMS  int64                   `json:"duration_ms"`
	Probability float64                 `json:"probability"`
	Amount      int                     `json:"amount"`
}

func (h *ChaosHandler) Run(c *gin.Context) {
	var req runExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	cfg := services.ExperimentConfig{
		Duration:    millisToDuration(req.DurationMS),
		Probability: req.Probability,
		Amount:      req.Amount,
		Target:      req.Target,
	}
	exp, err := h.svc.RunExperiment(c.Request.Context(), req.Type, cfg)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, exp)
}

func (h *ChaosHandler) List(c *gin.Context) {
	active := h.svc.GetActiveExperiments()
	history, _ := h.svc.GetExperimentHistory()
	c.JSON(http.StatusOK, gin.H{
		"active":  active,
		"history": history,
	})
}

func (h *ChaosHandler) Status(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		active := h.svc.GetActiveExperiments()
		c.JSON(http.StatusOK, gin.H{
			"active_count": len(active),
			"active":       active,
		})
		return
	}
	exp, ok := h.svc.GetExperimentStatus(id)
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "experiment not found", Code: "NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, exp)
}

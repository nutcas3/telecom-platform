package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// AutomationHandler exposes automation CRUD, run, and schedule endpoints.
type AutomationHandler struct {
	svc *services.AutomationService
}

// NewAutomationHandler constructs an AutomationHandler.
func NewAutomationHandler(svc *services.AutomationService) *AutomationHandler {
	return &AutomationHandler{svc: svc}
}

func (h *AutomationHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), services.AutomationFilter{
		Enabled: parseBoolQuery(c, "enabled"),
		Type:    c.Query("type"),
	})
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"automations": items, "total": len(items)})
}

func (h *AutomationHandler) Create(c *gin.Context) {
	var in services.CreateAutomationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AutomationHandler) Run(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.Run(c.Request.Context(), id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *AutomationHandler) Schedule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var in services.ScheduleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	a, err := h.svc.UpdateSchedule(c.Request.Context(), id, in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AutomationHandler) Logs(c *gin.Context) {
	var automationID uint
	if raw := c.Query("automation_id"); raw != "" {
		v, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			badRequest(c, err)
			return
		}
		automationID = uint(v)
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	runs, err := h.svc.ListRuns(c.Request.Context(), automationID, limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": runs, "total": len(runs)})
}

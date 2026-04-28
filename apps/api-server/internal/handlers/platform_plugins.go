package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// PluginsHandler exposes plugin install/lifecycle endpoints.
type PluginsHandler struct {
	svc *services.PluginService
}

// NewPluginsHandler constructs a PluginsHandler.
func NewPluginsHandler(svc *services.PluginService) *PluginsHandler {
	return &PluginsHandler{svc: svc}
}

func (h *PluginsHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), services.PluginFilter{
		Enabled:  parseBoolQuery(c, "enabled"),
		Type:     c.Query("type"),
		Category: c.Query("category"),
	})
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plugins": items, "total": len(items)})
}

func (h *PluginsHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	plugin, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, plugin)
}

func (h *PluginsHandler) Install(c *gin.Context) {
	var in services.InstallPluginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	p, err := h.svc.Install(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PluginsHandler) Uninstall(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Uninstall(c.Request.Context(), id); err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin uninstalled", "id": id})
}

func (h *PluginsHandler) Enable(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.SetEnabled(c.Request.Context(), id, true)
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PluginsHandler) Disable(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.SetEnabled(c.Request.Context(), id, false)
	if err != nil {
		notFound(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

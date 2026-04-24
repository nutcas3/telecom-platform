package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// ConfigHandler exposes runtime config read/write/validate endpoints.
type ConfigHandler struct {
	svc *services.ConfigStoreService
}

// NewConfigHandler constructs a ConfigHandler.
func NewConfigHandler(svc *services.ConfigStoreService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

func (h *ConfigHandler) Get(c *gin.Context) {
	entries, err := h.svc.List(c.Request.Context(), c.Query("section"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries, "total": len(entries)})
}

func (h *ConfigHandler) Update(c *gin.Context) {
	var in services.UpsertConfigInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, err)
		return
	}
	entry, err := h.svc.Upsert(c.Request.Context(), in)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *ConfigHandler) Validate(c *gin.Context) {
	res, err := h.svc.Validate(c.Request.Context(), c.Query("section"))
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

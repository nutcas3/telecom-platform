package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/whitelabel"
	"github.com/sirupsen/logrus"
)

// WhitelabelHandler handles whitelabel API requests
type WhitelabelHandler struct {
	service *whitelabel.Service
	logger  *logrus.Logger
}

// NewWhitelabelHandler creates a new whitelabel handler
func NewWhitelabelHandler(service *whitelabel.Service, logger *logrus.Logger) *WhitelabelHandler {
	return &WhitelabelHandler{service: service, logger: logger}
}

// CreateBranding creates branding configuration
func (h *WhitelabelHandler) CreateBranding(c *gin.Context) {
	var config whitelabel.BrandingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.CreateBranding(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetBranding retrieves branding configuration
func (h *WhitelabelHandler) GetBranding(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = c.Param("tenant_id")
	}

	config, err := h.service.GetBranding(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetBrandingByDomain retrieves branding by custom domain
func (h *WhitelabelHandler) GetBrandingByDomain(c *gin.Context) {
	domain := c.Param("domain")
	config, err := h.service.GetBrandingByDomain(c.Request.Context(), domain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateBranding updates branding configuration
func (h *WhitelabelHandler) UpdateBranding(c *gin.Context) {
	var config whitelabel.BrandingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.UpdateBranding(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CreatePartnerConfig creates partner configuration
func (h *WhitelabelHandler) CreatePartnerConfig(c *gin.Context) {
	var config whitelabel.PartnerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.CreatePartnerConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetPartnerConfig retrieves partner configuration
func (h *WhitelabelHandler) GetPartnerConfig(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	config, err := h.service.GetPartnerConfig(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CreateEmailTemplate creates an email template
func (h *WhitelabelHandler) CreateEmailTemplate(c *gin.Context) {
	var template whitelabel.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.CreateEmailTemplate(c.Request.Context(), &template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetEmailTemplate retrieves an email template
func (h *WhitelabelHandler) GetEmailTemplate(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	templateKey := c.Param("key")

	template, err := h.service.GetEmailTemplate(c.Request.Context(), tenantID, templateKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListEmailTemplates lists email templates
func (h *WhitelabelHandler) ListEmailTemplates(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	templates, err := h.service.ListEmailTemplates(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

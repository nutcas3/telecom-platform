package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/compliance"
	"github.com/sirupsen/logrus"
)

// ComplianceHandler handles compliance API requests
type ComplianceHandler struct {
	service *compliance.Service
	logger  *logrus.Logger
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(service *compliance.Service, logger *logrus.Logger) *ComplianceHandler {
	return &ComplianceHandler{service: service, logger: logger}
}

// CreateDSR creates a data subject request
func (h *ComplianceHandler) CreateDSR(c *gin.Context) {
	var req compliance.DataSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.CreateDSR(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetDSR retrieves a data subject request
func (h *ComplianceHandler) GetDSR(c *gin.Context) {
	id := c.Param("id")
	dsr, err := h.service.GetDSR(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dsr)
}

// ListDSRs lists data subject requests
func (h *ComplianceHandler) ListDSRs(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	var status *compliance.DSRStatus
	if s := c.Query("status"); s != "" {
		st := compliance.DSRStatus(s)
		status = &st
	}

	dsrs, err := h.service.ListDSRs(c.Request.Context(), tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dsrs)
}

// ProcessDSR processes a data subject request
func (h *ComplianceHandler) ProcessDSR(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.ProcessDSR(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processing"})
}

// RecordConsent records user consent
func (h *ComplianceHandler) RecordConsent(c *gin.Context) {
	var consent compliance.ConsentRecord
	if err := c.ShouldBindJSON(&consent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consent.TenantID = c.GetHeader("X-Tenant-ID")
	consent.IPAddress = c.ClientIP()
	consent.UserAgent = c.GetHeader("User-Agent")

	if err := h.service.RecordConsent(c.Request.Context(), &consent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, consent)
}

// RevokeConsent revokes user consent
func (h *ComplianceHandler) RevokeConsent(c *gin.Context) {
	subjectID := c.Param("subject_id")
	consentType := compliance.ConsentType(c.Query("type"))

	if err := h.service.RevokeConsent(c.Request.Context(), subjectID, consentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// GetConsents retrieves consent records for a subject
func (h *ComplianceHandler) GetConsents(c *gin.Context) {
	subjectID := c.Param("subject_id")
	consents, err := h.service.GetConsents(c.Request.Context(), subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, consents)
}

// QueryAuditLogs queries audit logs
func (h *ComplianceHandler) QueryAuditLogs(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	filter := &compliance.AuditLogFilter{
		Jurisdiction: c.Query("jurisdiction"),
		ActorID:      c.Query("actor_id"),
		Action:       c.Query("action"),
		Limit:        100,
	}

	logs, err := h.service.QueryAuditLogs(c.Request.Context(), tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// SetDataResidency sets data residency configuration
func (h *ComplianceHandler) SetDataResidency(c *gin.Context) {
	var config compliance.DataResidencyConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.TenantID = c.GetHeader("X-Tenant-ID")
	if err := h.service.SetDataResidency(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetDataResidency retrieves data residency configuration
func (h *ComplianceHandler) GetDataResidency(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	config, err := h.service.GetDataResidency(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

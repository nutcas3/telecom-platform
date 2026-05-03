package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService tenant.Service
	logger        *logrus.Logger
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService tenant.Service, logger *logrus.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// CreateTenant handles tenant creation requests
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req tenant.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind tenant creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create tenant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetTenant handles tenant retrieval requests
func (h *TenantHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	tenant, err := h.tenantService.GetTenant(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", id).Error("Failed to get tenant")
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetTenantByDomain handles tenant retrieval by domain requests
func (h *TenantHandler) GetTenantByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
		return
	}

	tenant, err := h.tenantService.GetTenantByDomain(c.Request.Context(), domain)
	if err != nil {
		h.logger.WithError(err).WithField("domain", domain).Error("Failed to get tenant by domain")
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant handles tenant update requests
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	var req tenant.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind tenant update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", id).Error("Failed to update tenant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant handles tenant deletion requests
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	if err := h.tenantService.DeleteTenant(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).WithField("tenant_id", id).Error("Failed to delete tenant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListTenants handles tenant listing requests
func (h *TenantHandler) ListTenants(c *gin.Context) {
	filter := &tenant.TenantFilter{
		Name:      c.Query("name"),
		Domain:    c.Query("domain"),
		Status:    tenant.TenantStatus(c.Query("status")),
		Plan:      tenant.TenantPlan(c.Query("plan")),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	tenants, err := h.tenantService.ListTenants(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tenants")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenants)
}

// AddUserToTenant handles user addition to tenant requests
func (h *TenantHandler) AddUserToTenant(c *gin.Context) {
	var req tenant.CreateTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind tenant user creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.tenantService.AddUserToTenant(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to add user to tenant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetTenantUser handles tenant user retrieval requests
func (h *TenantHandler) GetTenantUser(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	userID := c.Param("user_id")

	if tenantID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID and user ID are required"})
		return
	}

	user, err := h.tenantService.GetTenantUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"user_id":   userID,
		}).Error("Failed to get tenant user")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateTenantUser handles tenant user update requests
func (h *TenantHandler) UpdateTenantUser(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	userID := c.Param("user_id")

	if tenantID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID and user ID are required"})
		return
	}

	var req tenant.UpdateTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind tenant user update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.tenantService.UpdateTenantUser(c.Request.Context(), tenantID, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"user_id":   userID,
		}).Error("Failed to update tenant user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// RemoveUserFromTenant handles user removal from tenant requests
func (h *TenantHandler) RemoveUserFromTenant(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	userID := c.Param("user_id")

	if tenantID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID and user ID are required"})
		return
	}

	if err := h.tenantService.RemoveUserFromTenant(c.Request.Context(), tenantID, userID); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"user_id":   userID,
		}).Error("Failed to remove user from tenant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListTenantUsers handles tenant user listing requests
func (h *TenantHandler) ListTenantUsers(c *gin.Context) {
	filter := &tenant.TenantUserFilter{
		TenantID: c.Param("tenant_id"),
		Email:    c.Query("email"),
		Role:     tenant.TenantRole(c.Query("role")),
		Status:   tenant.TenantUserStatus(c.Query("status")),
	}

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	users, err := h.tenantService.ListTenantUsers(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tenant users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// CreateAPIKey handles API key creation requests
func (h *TenantHandler) CreateAPIKey(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	var req tenant.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind API key creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, keyString, err := h.tenantService.CreateAPIKey(c.Request.Context(), tenantID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to create API key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return API key and the actual key (only shown once)
	response := gin.H{
		"api_key": apiKey,
		"key":     keyString, // Only returned on creation
	}

	c.JSON(http.StatusCreated, response)
}

// GetAPIKey handles API key retrieval requests
func (h *TenantHandler) GetAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	apiKey, err := h.tenantService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("key_id", id).Error("Failed to get API key")
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// UpdateAPIKey handles API key update requests
func (h *TenantHandler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	var req tenant.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind API key update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.tenantService.UpdateAPIKey(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithError(err).WithField("key_id", id).Error("Failed to update API key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// DeleteAPIKey handles API key deletion requests
func (h *TenantHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("key_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	if err := h.tenantService.DeleteAPIKey(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).WithField("key_id", id).Error("Failed to delete API key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListAPIKeys handles API key listing requests
func (h *TenantHandler) ListAPIKeys(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	apiKeys, err := h.tenantService.ListAPIKeys(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to list API keys")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

// GetUsageStats handles usage statistics requests
func (h *TenantHandler) GetUsageStats(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	stats, err := h.tenantService.GetUsageStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get usage stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetQuotaStatus handles quota status requests
func (h *TenantHandler) GetQuotaStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	quotaStatus, err := h.tenantService.GetQuotaStatus(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get quota status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quotaStatus)
}

// GetTenantConfig handles tenant configuration requests
func (h *TenantHandler) GetTenantConfig(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	config, err := h.tenantService.GetTenantConfig(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateTenantConfig handles tenant configuration update requests
func (h *TenantHandler) UpdateTenantConfig(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	var config tenant.TenantConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.WithError(err).Error("Failed to bind tenant config request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantService.UpdateTenantConfig(c.Request.Context(), tenantID, &config); err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to update tenant config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetTenantMetrics handles tenant metrics requests
func (h *TenantHandler) GetTenantMetrics(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	metrics, err := h.tenantService.GetTenantMetrics(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTenantEvents handles tenant events requests
func (h *TenantHandler) GetTenantEvents(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	limit := 100 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.tenantService.GetTenantEvents(c.Request.Context(), tenantID, limit)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

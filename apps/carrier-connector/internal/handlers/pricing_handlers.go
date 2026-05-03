package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
)

// CreateRule creates a new pricing rule
func (h *PricingHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context (assuming middleware sets it)
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
		return
	}

	rule := &pricing.PricingRule{
		ID:          id.GeneratePrefixed("rule"),
		Name:        req.Name,
		Description: req.Description,
		TenantID:    tenantID,
		Type:        req.Type,
		Priority:    req.Priority,
		IsActive:    true,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		Metadata:    req.Metadata,
	}

	createdRule, err := h.service.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdRule)
}

// GetRule retrieves a pricing rule
func (h *PricingHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID required"})
		return
	}

	rule, err := h.service.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateRule updates a pricing rule
func (h *PricingHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID required"})
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing rule
	rule, err := h.service.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}

	// Apply updates
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Type != nil {
		rule.Type = *req.Type
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Conditions != nil {
		rule.Conditions = *req.Conditions
	}
	if req.Actions != nil {
		rule.Actions = *req.Actions
	}
	if req.Metadata != nil {
		rule.Metadata = req.Metadata
	}

	updatedRule, err := h.service.UpdateRule(c.Request.Context(), id, rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedRule)
}

// DeleteRule deletes a pricing rule
func (h *PricingHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule ID required"})
		return
	}

	err := h.service.DeleteRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule deleted successfully"})
}

// ListRules lists pricing rules
func (h *PricingHandler) ListRules(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
		return
	}

	// Parse query parameters
	filter := &pricing.PricingFilter{
		TenantID: tenantID,
	}

	if ruleType := c.Query("type"); ruleType != "" {
		filter.Type = ruleType
	}

	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filter.IsActive = &active
		}
	}

	if priority := c.Query("priority"); priority != "" {
		if prio, err := strconv.Atoi(priority); err == nil {
			filter.Priority = &prio
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if lim, err := strconv.Atoi(limit); err == nil {
			filter.Limit = lim
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if off, err := strconv.Atoi(offset); err == nil {
			filter.Offset = off
		}
	}

	rules, err := h.service.ListRules(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

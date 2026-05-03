package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/pricing"
)

// PricingHandler handles pricing-related HTTP requests
type PricingHandler struct {
	service pricing.Service
}

// NewPricingHandler creates a new pricing handler
func NewPricingHandler(service pricing.Service) *PricingHandler {
	return &PricingHandler{service: service}
}

// RegisterRoutes registers pricing routes
func (h *PricingHandler) RegisterRoutes(router *gin.RouterGroup) {
	pricing := router.Group("/pricing")
	{
		// Rule management
		pricing.POST("/rules", h.CreateRule)
		pricing.GET("/rules/:id", h.GetRule)
		pricing.PUT("/rules/:id", h.UpdateRule)
		pricing.DELETE("/rules/:id", h.DeleteRule)
		pricing.GET("/rules", h.ListRules)

		// Pricing calculations
		pricing.POST("/calculate", h.CalculatePrice)
		pricing.POST("/apply-rules", h.ApplyRules)

		// Analytics
		pricing.GET("/analytics", h.GetAnalytics)
	}
}

type CreateRuleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Type        pricing.RuleType       `json:"type" binding:"required"`
	Priority    int                    `json:"priority"`
	Conditions  pricing.RuleConditions `json:"conditions"`
	Actions     pricing.RuleActions    `json:"actions"`
	Metadata    map[string]any         `json:"metadata"`
}

type UpdateRuleRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Type        *pricing.RuleType       `json:"type,omitempty"`
	Priority    *int                    `json:"priority,omitempty"`
	IsActive    *bool                   `json:"is_active,omitempty"`
	Conditions  *pricing.RuleConditions `json:"conditions,omitempty"`
	Actions     *pricing.RuleActions    `json:"actions,omitempty"`
	Metadata    map[string]any          `json:"metadata,omitempty"`
}

type CalculatePriceRequest struct {
	TenantID   string         `json:"tenant_id" binding:"required"`
	CustomerID string         `json:"customer_id" binding:"required"`
	ProductID  string         `json:"product_id" binding:"required"`
	BasePrice  float64        `json:"base_price" binding:"required"`
	Currency   string         `json:"currency" binding:"required"`
	Quantity   int            `json:"quantity" binding:"required"`
	Location   string         `json:"location"`
	Metadata   map[string]any `json:"metadata"`
}

type ApplyRulesRequest struct {
	Context pricing.PricingContext `json:"context" binding:"required"`
	Rules   []*pricing.PricingRule `json:"rules"`
}

// ApplyRules applies specific rules
func (h *PricingHandler) ApplyRules(c *gin.Context) {
	var req ApplyRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.ApplyRules(c.Request.Context(), &req.Context, req.Rules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAnalytics retrieves pricing analytics
func (h *PricingHandler) GetAnalytics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID required"})
		return
	}

	analytics, err := h.service.GetAnalytics(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}


// CalculatePrice calculates price based on active rules
func (h *PricingHandler) CalculatePrice(c *gin.Context) {
	var req CalculatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	context := &pricing.PricingContext{
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		BasePrice:  req.BasePrice,
		Currency:   req.Currency,
		Quantity:   req.Quantity,
		Location:   req.Location,
		Time:       time.Now(),
		Metadata:   req.Metadata,
	}

	result, err := h.service.CalculatePrice(c.Request.Context(), context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}


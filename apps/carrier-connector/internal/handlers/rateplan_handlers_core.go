package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/rateplan"
)

// RatePlanHandler handles rate plan API endpoints
type RatePlanHandler struct {
	service rateplan.Service
	logger  *logrus.Logger
}

// NewRatePlanHandler creates a new rate plan handler
func NewRatePlanHandler(service rateplan.Service) *RatePlanHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &RatePlanHandler{
		service: service,
		logger:  logger,
	}
}

// CreateRatePlanRequest represents the request to create a rate plan
type CreateRatePlanRequest struct {
	Name             string                     `json:"name" binding:"required"`
	Description      string                     `json:"description"`
	CarrierID        string                     `json:"carrier_id" binding:"required"`
	Region           string                     `json:"region" binding:"required"`
	PlanType         rateplan.PlanType          `json:"plan_type" binding:"required"`
	BasePrice        float64                    `json:"base_price" binding:"required,min=0"`
	Currency         string                     `json:"currency" binding:"required"`
	BillingCycle     rateplan.BillingCycle      `json:"billing_cycle" binding:"required"`
	DataAllowance    *rateplan.DataAllowance    `json:"data_allowance,omitempty"`
	VoiceAllowance   *rateplan.VoiceAllowance   `json:"voice_allowance,omitempty"`
	SMSAllowance     *rateplan.SMSAllowance     `json:"sms_allowance,omitempty"`
	OverageRates     *rateplan.OverageRates     `json:"overage_rates,omitempty"`
	Features         []rateplan.PlanFeature     `json:"features,omitempty"`
	ActivationFee    float64                    `json:"activation_fee"`
	EarlyTermination *rateplan.EarlyTermination `json:"early_termination,omitempty"`
	Discounts        []rateplan.Discount        `json:"discounts,omitempty"`
	ValidFrom        time.Time                  `json:"valid_from" binding:"required"`
	ValidTo          *time.Time                 `json:"valid_to,omitempty"`
	Priority         int                        `json:"priority"`
	IsActive         bool                       `json:"is_active"`
	Metadata         map[string]any             `json:"metadata,omitempty"`
}

// UpdateRatePlanRequest represents the request to update a rate plan
type UpdateRatePlanRequest struct {
	Name             string                     `json:"name,omitempty"`
	Description      string                     `json:"description,omitempty"`
	PlanType         rateplan.PlanType          `json:"plan_type,omitempty"`
	Status           rateplan.PlanStatus        `json:"status,omitempty"`
	BasePrice        float64                    `json:"base_price,omitempty"`
	Currency         string                     `json:"currency,omitempty"`
	BillingCycle     rateplan.BillingCycle      `json:"billing_cycle,omitempty"`
	DataAllowance    *rateplan.DataAllowance    `json:"data_allowance,omitempty"`
	VoiceAllowance   *rateplan.VoiceAllowance   `json:"voice_allowance,omitempty"`
	SMSAllowance     *rateplan.SMSAllowance     `json:"sms_allowance,omitempty"`
	OverageRates     *rateplan.OverageRates     `json:"overage_rates,omitempty"`
	Features         []rateplan.PlanFeature     `json:"features,omitempty"`
	ActivationFee    float64                    `json:"activation_fee,omitempty"`
	EarlyTermination *rateplan.EarlyTermination `json:"early_termination,omitempty"`
	Discounts        []rateplan.Discount        `json:"discounts,omitempty"`
	ValidFrom        *time.Time                 `json:"valid_from,omitempty"`
	ValidTo          *time.Time                 `json:"valid_to,omitempty"`
	Priority         int                        `json:"priority,omitempty"`
	IsActive         *bool                      `json:"is_active,omitempty"`
	Metadata         map[string]any             `json:"metadata,omitempty"`
}

// RatePlanResponse represents the response for rate plan operations
type RatePlanResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message,omitempty"`
	Data    *rateplan.RatePlan `json:"data,omitempty"`
}

// RatePlansResponse represents the response for listing rate plans
type RatePlansResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message,omitempty"`
	Data    []*rateplan.RatePlan `json:"data,omitempty"`
	Total   int                  `json:"total,omitempty"`
}

// writeJSONResponse writes a JSON response
func (h *RatePlanHandler) writeJSONResponse(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, data)
}

// writeErrorResponse writes an error response
func (h *RatePlanHandler) writeErrorResponse(c *gin.Context, statusCode int, message string) {
	response := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{
		Success: false,
		Error:   message,
	}

	h.writeJSONResponse(c, statusCode, response)
}

// extractIDFromPath extracts ID from URL path
func (h *RatePlanHandler) extractIDFromPath(c *gin.Context) string {
	return c.Param("id")
}

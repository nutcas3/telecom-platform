package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
)

// PaymentController handles payment-related HTTP endpoints.
type PaymentController struct {
	paymentService *PaymentService
	db             *database.Database
}

// NewPaymentController creates a new payment controller.
func NewPaymentController(paymentService *PaymentService, db *database.Database) *PaymentController {
	return &PaymentController{paymentService: paymentService, db: db}
}

// SetupRoutes registers payment and webhook routes on the given group.
func (pc *PaymentController) SetupRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.POST("/top-up", pc.ProcessTopUp)
		payments.POST("/payment-methods", pc.CreatePaymentMethod)
		payments.DELETE("/payment-methods/:id", pc.DeletePaymentMethod)
		payments.GET("/payment-methods", pc.ListPaymentMethods)
		payments.GET("/transactions", pc.ListTransactions)
		payments.GET("/transactions/:id", pc.GetTransaction)
		payments.POST("/transactions/:id/refund", pc.RefundTransaction)
	}
	webhook := router.Group("/webhooks")
	{
		webhook.POST("/stripe", pc.HandleStripeWebhook)
	}
}

// ProcessTopUp processes a top-up payment.
func (pc *PaymentController) ProcessTopUp(c *gin.Context) {
	var req struct {
		Amount          float64 `json:"amount" binding:"required,min=0.01"`
		PaymentMethodID string  `json:"paymentMethodId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transaction, err := pc.paymentService.ProcessTopUp(c.Request.Context(), subscriberID, req.Amount, req.PaymentMethodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction": transaction,
		"message":     "Top-up processed successfully",
	})
}

// CreatePaymentMethod creates a new payment method.
func (pc *PaymentController) CreatePaymentMethod(c *gin.Context) {
	var req struct {
		Type      string            `json:"type" binding:"required"`
		Token     string            `json:"token" binding:"required"`
		IsDefault bool              `json:"isDefault"`
		Metadata  map[string]string `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	createReq := &CreatePaymentMethodRequest{
		Type:      PaymentMethodType(req.Type),
		Token:     req.Token,
		IsDefault: req.IsDefault,
		Metadata:  req.Metadata,
	}

	paymentMethod, err := pc.paymentService.CreatePaymentMethod(c.Request.Context(), subscriberID, createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"paymentMethod": paymentMethod,
		"message":       "Payment method created successfully",
	})
}

// DeletePaymentMethod deletes a payment method, enforcing ownership.
func (pc *PaymentController) DeletePaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	if paymentMethodID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment method ID required"})
		return
	}

	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pm, err := pc.db.GetPaymentMethod(c.Request.Context(), paymentMethodID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	if pm.SubscriberID != uint(subscriberID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Payment method does not belong to subscriber"})
		return
	}

	if err := pc.paymentService.gateway.DeletePaymentMethod(c.Request.Context(), pm.GatewayID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := pc.db.DeletePaymentMethod(c.Request.Context(), paymentMethodID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted successfully"})
}

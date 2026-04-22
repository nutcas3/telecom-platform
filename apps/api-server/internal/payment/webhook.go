package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// WebhookEvent represents a payment webhook event
type WebhookEvent struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Object   string         `json:"object"`
	Created  int64          `json:"created"`
	Livemode bool           `json:"livemode"`
	Data     map[string]any `json:"data"`
}

// WebhookHandler handles payment webhooks
type WebhookHandler struct {
	paymentService *PaymentService
	db             *database.Database
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(paymentService *PaymentService, db *database.Database) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
		db:             db,
	}
}

// HandleStripeWebhook handles Stripe webhooks
func (wh *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	// Validate webhook
	valid, err := wh.paymentService.ValidateWebhook(c.Request.Context(), signature, body)
	if err != nil {
		log.Printf("Webhook validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook"})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse event
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to parse webhook event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	// Handle different event types
	switch event.Type {
	case "payment_intent.succeeded":
		err = wh.handlePaymentSucceeded(c.Request.Context(), event)
	case "payment_intent.payment_failed":
		err = wh.handlePaymentFailed(c.Request.Context(), event)
	case "payment_intent.canceled":
		err = wh.handlePaymentCanceled(c.Request.Context(), event)
	case "charge.dispute.created":
		err = wh.handleDisputeCreated(c.Request.Context(), event)
	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	if err != nil {
		log.Printf("Error handling webhook event %s: %v", event.Type, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// handlePaymentSucceeded handles successful payment events
func (wh *WebhookHandler) handlePaymentSucceeded(ctx context.Context, event WebhookEvent) error {
	paymentIntent, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payment intent object")
	}

	transactionID, ok := paymentIntent["id"].(string)
	if !ok {
		return fmt.Errorf("missing transaction ID")
	}

	amount := float64(paymentIntent["amount"].(int64)) / 100
	currency := paymentIntent["currency"].(string)

	// Get transaction from database
	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update transaction status
	transaction.Status = "COMPLETED"
	transaction.UpdatedAt = time.Now()

	err = wh.db.UpdateTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// If it's a top-up, update subscriber balance
	if transaction.Type == "TOP_UP" {
		err = wh.db.UpdateSubscriberBalance(ctx, transaction.SubscriberID, amount)
		if err != nil {
			return fmt.Errorf("failed to update subscriber balance: %w", err)
		}

		// Send notification to subscriber
		err = wh.sendTopUpNotification(ctx, transaction.SubscriberID, amount, currency)
		if err != nil {
			log.Printf("Failed to send top-up notification: %v", err)
		}
	}

	log.Printf("Payment succeeded: %s, amount: %.2f %s", transactionID, amount, currency)
	return nil
}

// handlePaymentFailed handles failed payment events
func (wh *WebhookHandler) handlePaymentFailed(ctx context.Context, event WebhookEvent) error {
	paymentIntent, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payment intent object")
	}

	transactionID, ok := paymentIntent["id"].(string)
	if !ok {
		return fmt.Errorf("missing transaction ID")
	}

	// Get transaction from database
	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update transaction status
	transaction.Status = "FAILED"
	transaction.UpdatedAt = time.Now()

	// Add failure reason if available
	if lastPaymentError, ok := paymentIntent["last_payment_error"].(map[string]any); ok {
		if message, ok := lastPaymentError["message"].(string); ok {
			transaction.Description += " - Failed: " + message
		}
	}

	err = wh.db.UpdateTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Create alert for failed payment
	subID := int(transaction.SubscriberID)
	alert := &models.Alert{
		SubscriberID: &subID,
		Type:         models.AlertTypePaymentFailed,
		Severity:     models.AlertSeverityMedium,
		Message:      fmt.Sprintf("Payment of %.2f %s failed", transaction.Amount, transaction.Currency),
		Resolved:     false,
		Timestamp:    time.Now(),
	}

	err = wh.db.CreateAlert(ctx, alert)
	if err != nil {
		log.Printf("Failed to create payment failure alert: %v", err)
	}

	log.Printf("Payment failed: %s", transactionID)
	return nil
}

// handlePaymentCanceled handles canceled payment events
func (wh *WebhookHandler) handlePaymentCanceled(ctx context.Context, event WebhookEvent) error {
	paymentIntent, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payment intent object")
	}

	transactionID, ok := paymentIntent["id"].(string)
	if !ok {
		return fmt.Errorf("missing transaction ID")
	}

	// Get transaction from database
	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update transaction status
	transaction.Status = "CANCELLED"
	transaction.UpdatedAt = time.Now()

	err = wh.db.UpdateTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	log.Printf("Payment canceled: %s", transactionID)
	return nil
}

// handleDisputeCreated handles dispute events
func (wh *WebhookHandler) handleDisputeCreated(ctx context.Context, event WebhookEvent) error {
	dispute, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid dispute object")
	}

	chargeID, ok := dispute["charge"].(string)
	if !ok {
		return fmt.Errorf("missing charge ID")
	}

	amount := float64(dispute["amount"].(int64)) / 100
	reason, _ := dispute["reason"].(string)

	// Find transaction by charge ID
	transaction, err := wh.db.GetTransactionByChargeID(ctx, chargeID)
	if err != nil {
		return fmt.Errorf("failed to find transaction for charge %s: %w", chargeID, err)
	}

	// Create high-priority alert for dispute
	dispSubID := int(transaction.SubscriberID)
	alert := &models.Alert{
		SubscriberID: &dispSubID,
		Type:         models.AlertTypePaymentFailed,
		Severity:     models.AlertSeverityHigh,
		Message:      fmt.Sprintf("Payment dispute created for %.2f %s. Reason: %s", amount, transaction.Currency, reason),
		Resolved:     false,
		Timestamp:    time.Now(),
	}

	err = wh.db.CreateAlert(ctx, alert)
	if err != nil {
		return fmt.Errorf("failed to create dispute alert: %w", err)
	}

	log.Printf("Dispute created for charge %s, amount: %.2f", chargeID, amount)
	return nil
}

// sendTopUpNotification sends a notification for successful top-up
func (wh *WebhookHandler) sendTopUpNotification(ctx context.Context, subscriberID uint, amount float64, currency string) error {
	// Get subscriber details
	subscriber, err := wh.db.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Create notification
	notification := &models.Notification{
		SubscriberID: subscriberID,
		Type:         "PAYMENT_SUCCESS",
		Title:        "Top-up Successful",
		Message: fmt.Sprintf("Your account has been topped up with %.2f %s. New balance: %.2f %s",
			amount, currency, subscriber.Balance, currency),
		Read:      false,
		CreatedAt: time.Now(),
	}

	err = wh.db.CreateNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send email notification (implementation would depend on email service)
	// This is a placeholder for email sending logic
	log.Printf("Top-up notification sent to subscriber %d: %.2f %s", subscriberID, amount, currency)

	return nil
}

// PaymentController handles payment-related HTTP endpoints
type PaymentController struct {
	paymentService *PaymentService
	db             *database.Database
}

// NewPaymentController creates a new payment controller
func NewPaymentController(paymentService *PaymentService, db *database.Database) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
		db:             db,
	}
}

// SetupRoutes sets up payment routes
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

// ProcessTopUp processes a top-up payment
func (pc *PaymentController) ProcessTopUp(c *gin.Context) {
	var req struct {
		Amount          float64 `json:"amount" binding:"required,min=0.01"`
		PaymentMethodID string  `json:"paymentMethodId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get subscriber ID from context (would be set by auth middleware)
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

// CreatePaymentMethod creates a new payment method
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

// DeletePaymentMethod deletes a payment method
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

	// Verify payment method belongs to subscriber
	pm, err := pc.db.GetPaymentMethod(c.Request.Context(), paymentMethodID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	if pm.SubscriberID != uint(subscriberID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Payment method does not belong to subscriber"})
		return
	}

	err = pc.paymentService.gateway.DeletePaymentMethod(c.Request.Context(), pm.GatewayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete from database
	err = pc.db.DeletePaymentMethod(c.Request.Context(), paymentMethodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted successfully"})
}

// ListPaymentMethods lists payment methods for a subscriber
func (pc *PaymentController) ListPaymentMethods(c *gin.Context) {
	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	paymentMethods, err := pc.db.ListPaymentMethods(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"paymentMethods": paymentMethods})
}

// ListTransactions lists transactions for a subscriber
func (pc *PaymentController) ListTransactions(c *gin.Context) {
	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transactions, err := pc.db.ListTransactions(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetTransaction gets a specific transaction
func (pc *PaymentController) GetTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID required"})
		return
	}

	subscriberID := c.GetInt("subscriberID")
	if subscriberID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transaction, err := pc.db.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if transaction.SubscriberID != uint(subscriberID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transaction does not belong to subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

// RefundTransaction refunds a transaction
func (pc *PaymentController) RefundTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID required"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0.01"`
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

	// Get transaction
	transaction, err := pc.db.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if transaction.SubscriberID != uint(subscriberID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Transaction does not belong to subscriber"})
		return
	}

	if transaction.Status != "COMPLETED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot refund non-completed transaction"})
		return
	}

	// Process refund
	refundResp, err := pc.paymentService.gateway.RefundPayment(c.Request.Context(), transaction.TransactionID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create refund record
	refund := &models.Transaction{
		SubscriberID:  transaction.SubscriberID,
		TransactionID: refundResp.RefundID,
		Type:          "REFUND",
		Amount:        req.Amount,
		Currency:      refundResp.Currency,
		Status:        string(refundResp.Status),
		Description:   fmt.Sprintf("Refund for transaction %s", transactionID),
		ParentID:      &transaction.ID,
		CreatedAt:     refundResp.ProcessedAt,
		UpdatedAt:     refundResp.ProcessedAt,
	}

	err = pc.db.CreateTransaction(c.Request.Context(), refund)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refund"})
		return
	}

	// Update subscriber balance
	err = pc.db.UpdateSubscriberBalance(c.Request.Context(), transaction.SubscriberID, -req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund":  refund,
		"message": "Refund processed successfully",
	})
}

// HandleStripeWebhook handles Stripe webhooks (delegates to webhook handler)
func (pc *PaymentController) HandleStripeWebhook(c *gin.Context) {
	webhookHandler := NewWebhookHandler(pc.paymentService, pc.db)
	webhookHandler.HandleStripeWebhook(c)
}

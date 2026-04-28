package payment

import (
"encoding/json"
"log"
"net/http"

"github.com/gin-gonic/gin"

"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
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


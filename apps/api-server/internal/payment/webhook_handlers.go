package payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// handlePaymentSucceeded handles successful payment events.
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

	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction.Status = "COMPLETED"
	transaction.UpdatedAt = time.Now()
	if err := wh.db.UpdateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	if transaction.Type == "TOP_UP" {
		if err := wh.db.UpdateSubscriberBalance(ctx, transaction.SubscriberID, amount); err != nil {
			return fmt.Errorf("failed to update subscriber balance: %w", err)
		}
		if err := wh.sendTopUpNotification(ctx, transaction.SubscriberID, amount, currency); err != nil {
			log.Printf("Failed to send top-up notification: %v", err)
		}
	}

	log.Printf("Payment succeeded: %s, amount: %.2f %s", transactionID, amount, currency)
	return nil
}

// handlePaymentFailed handles failed payment events.
func (wh *WebhookHandler) handlePaymentFailed(ctx context.Context, event WebhookEvent) error {
	paymentIntent, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payment intent object")
	}
	transactionID, ok := paymentIntent["id"].(string)
	if !ok {
		return fmt.Errorf("missing transaction ID")
	}

	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction.Status = "FAILED"
	transaction.UpdatedAt = time.Now()
	if lastErr, ok := paymentIntent["last_payment_error"].(map[string]any); ok {
		if message, ok := lastErr["message"].(string); ok {
			transaction.Description += " - Failed: " + message
		}
	}
	if err := wh.db.UpdateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	subID := int(transaction.SubscriberID)
	alert := &models.Alert{
		SubscriberID: &subID,
		Type:         models.AlertTypePaymentFailed,
		Severity:     models.AlertSeverityMedium,
		Message:      fmt.Sprintf("Payment of %.2f %s failed", transaction.Amount, transaction.Currency),
		Resolved:     false,
		Timestamp:    time.Now(),
	}
	if err := wh.db.CreateAlert(ctx, alert); err != nil {
		log.Printf("Failed to create payment failure alert: %v", err)
	}

	log.Printf("Payment failed: %s", transactionID)
	return nil
}

// handlePaymentCanceled handles canceled payment events.
func (wh *WebhookHandler) handlePaymentCanceled(ctx context.Context, event WebhookEvent) error {
	paymentIntent, ok := event.Data["object"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payment intent object")
	}
	transactionID, ok := paymentIntent["id"].(string)
	if !ok {
		return fmt.Errorf("missing transaction ID")
	}

	transaction, err := wh.db.GetTransactionByGatewayID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	transaction.Status = "CANCELLED"
	transaction.UpdatedAt = time.Now()
	if err := wh.db.UpdateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	log.Printf("Payment canceled: %s", transactionID)
	return nil
}

// handleDisputeCreated handles dispute events.
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

	transaction, err := wh.db.GetTransactionByChargeID(ctx, chargeID)
	if err != nil {
		return fmt.Errorf("failed to find transaction for charge %s: %w", chargeID, err)
	}

	dispSubID := int(transaction.SubscriberID)
	alert := &models.Alert{
		SubscriberID: &dispSubID,
		Type:         models.AlertTypePaymentFailed,
		Severity:     models.AlertSeverityHigh,
		Message:      fmt.Sprintf("Payment dispute created for %.2f %s. Reason: %s", amount, transaction.Currency, reason),
		Resolved:     false,
		Timestamp:    time.Now(),
	}
	if err := wh.db.CreateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create dispute alert: %w", err)
	}

	log.Printf("Dispute created for charge %s, amount: %.2f", chargeID, amount)
	return nil
}

// sendTopUpNotification sends a notification for successful top-up.
func (wh *WebhookHandler) sendTopUpNotification(ctx context.Context, subscriberID uint, amount float64, currency string) error {
	subscriber, err := wh.db.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	notification := &models.Notification{
		SubscriberID: subscriberID,
		Type:         "PAYMENT_SUCCESS",
		Title:        "Top-up Successful",
		Message: fmt.Sprintf("Your account has been topped up with %.2f %s. New balance: %.2f %s",
			amount, currency, subscriber.Balance, currency),
		Read:      false,
		CreatedAt: time.Now(),
	}
	if err := wh.db.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("Top-up notification sent to subscriber %d: %.2f %s", subscriberID, amount, currency)
	return nil
}

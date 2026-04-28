package payment

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
)

// StripeGateway implements PaymentGateway for Stripe.
type StripeGateway struct {
	apiKey        string
	webhookSecret string
	client        *http.Client
	config        *config.PaymentConfig
}

// NewStripeGateway creates a new Stripe payment gateway.
func NewStripeGateway(cfg *config.PaymentConfig) *StripeGateway {
	return &StripeGateway{
		apiKey:        cfg.StripeAPIKey,
		webhookSecret: cfg.StripeWebhookSecret,
		client:        &http.Client{Timeout: 30 * time.Second},
		config:        cfg,
	}
}

// ProcessPayment processes a payment via Stripe.
func (s *StripeGateway) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	intentData := map[string]any{
		"amount":              int64(req.Amount * 100),
		"currency":            strings.ToLower(req.Currency),
		"description":         req.Description,
		"customer":            req.CustomerID,
		"payment_method":      req.PaymentMethodID,
		"confirmation_method": "manual",
		"confirm":             true,
		"metadata":            req.Metadata,
	}

	if req.WebhookURL != "" {
		intentData["transfer_data"] = map[string]string{"destination": req.WebhookURL}
	}

	intentResp, err := s.makeStripeRequest(ctx, "POST", "/v1/payment_intents", intentData)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}
	intent, ok := intentResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payment intent response")
	}

	status := s.mapStripeStatus(intent["status"].(string))

	var failureReason *string
	if status == PaymentStatusFailed {
		if lastErr, ok := intent["last_payment_error"].(map[string]any); ok {
			if message, ok := lastErr["message"].(string); ok {
				failureReason = &message
			}
		}
	}

	var redirectURL *string
	if status == PaymentStatusPending {
		if nextAction, ok := intent["next_action"].(map[string]any); ok {
			if url, ok := nextAction["redirect_to_url"].(map[string]any); ok {
				if redirect, ok := url["url"].(string); ok {
					redirectURL = &redirect
				}
			}
		}
	}

	return &PaymentResponse{
		TransactionID: intent["id"].(string),
		Status:        status,
		Amount:        req.Amount,
		Currency:      req.Currency,
		ProcessedAt:   time.Now(),
		FailureReason: failureReason,
		Metadata:      req.Metadata,
		RedirectURL:   redirectURL,
	}, nil
}

// RefundPayment refunds a payment via Stripe.
func (s *StripeGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResponse, error) {
	refundData := map[string]any{
		"payment_intent": transactionID,
		"amount":         int64(amount * 100),
	}
	refundResp, err := s.makeStripeRequest(ctx, "POST", "/v1/refunds", refundData)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}
	refund, ok := refundResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid refund response")
	}

	status := s.mapStripeStatus(refund["status"].(string))

	var failureReason *string
	if status == PaymentStatusFailed {
		if failureMsg, ok := refund["failure_reason"].(string); ok && failureMsg != "" {
			failureReason = &failureMsg
		}
	}

	return &RefundResponse{
		RefundID:      refund["id"].(string),
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      "USD",
		Status:        status,
		ProcessedAt:   time.Now(),
		FailureReason: failureReason,
	}, nil
}

// GetPaymentStatus gets payment status from Stripe.
func (s *StripeGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	intentResp, err := s.makeStripeRequest(ctx, "GET", "/v1/payment_intents/"+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}
	intent, ok := intentResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payment intent response")
	}

	refundsResp, err := s.makeStripeRequest(ctx, "GET", "/v1/refunds?payment_intent="+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get refunds: %w", err)
	}
	refundsData, _ := refundsResp.(map[string]any)
	refundsList, _ := refundsData["data"].([]any)

	refunds := make([]RefundInfo, len(refundsList))
	for i, refundInterface := range refundsList {
		refund, _ := refundInterface.(map[string]any)
		createdAt, _ := time.Parse(time.RFC3339, refund["created"].(string))
		refunds[i] = RefundInfo{
			RefundID:  refund["id"].(string),
			Amount:    float64(refund["amount"].(int64)) / 100,
			Status:    refund["status"].(string),
			CreatedAt: createdAt,
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, intent["created"].(string))

	var failureReason *string
	if lastErr, ok := intent["last_payment_error"].(map[string]any); ok {
		if message, ok := lastErr["message"].(string); ok {
			failureReason = &message
		}
	}

	return &PaymentStatus{
		TransactionID: transactionID,
		Status:        s.mapStripeStatus(intent["status"].(string)),
		Amount:        float64(intent["amount"].(int64)) / 100,
		Currency:      strings.ToUpper(intent["currency"].(string)),
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
		FailureReason: failureReason,
		Refunds:       refunds,
	}, nil
}

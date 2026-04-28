package payment

import (
	"context"
	"crypto/hmac"
	"fmt"
	"strings"
	"time"
)

// CreatePaymentMethod creates a payment method in Stripe and attaches it to the customer.
func (s *StripeGateway) CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*PaymentMethod, error) {
	pmData := map[string]any{
		"type":     s.mapPaymentMethodType(req.Type),
		"card":     map[string]string{"token": req.Token},
		"metadata": req.Metadata,
	}

	pmResp, err := s.makeStripeRequest(ctx, "POST", "/v1/payment_methods", pmData)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}
	pm, ok := pmResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payment method response")
	}

	attachData := map[string]any{"customer": req.CustomerID}
	if _, err := s.makeStripeRequest(ctx, "POST", "/v1/payment_methods/"+pm["id"].(string)+"/attach", attachData); err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	if req.IsDefault {
		customerData := map[string]any{
			"invoice_settings": map[string]string{"default_payment_method": pm["id"].(string)},
		}
		if _, err := s.makeStripeRequest(ctx, "POST", "/v1/customers/"+req.CustomerID, customerData); err != nil {
			return nil, fmt.Errorf("failed to set default payment method: %w", err)
		}
	}

	card, _ := pm["card"].(map[string]any)
	createdAt, _ := time.Parse(time.RFC3339, pm["created"].(string))

	return &PaymentMethod{
		ID:          pm["id"].(string),
		Type:        req.Type,
		CustomerID:  req.CustomerID,
		Last4:       card["last4"].(string),
		Brand:       card["brand"].(string),
		ExpiryMonth: int(card["exp_month"].(float64)),
		ExpiryYear:  int(card["exp_year"].(float64)),
		IsDefault:   req.IsDefault,
		CreatedAt:   createdAt,
		Metadata:    req.Metadata,
	}, nil
}

// DeletePaymentMethod deletes (detaches) a payment method from Stripe.
func (s *StripeGateway) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	if _, err := s.makeStripeRequest(ctx, "DELETE", "/v1/payment_methods/"+paymentMethodID, nil); err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}
	return nil
}

// ValidateWebhook validates a Stripe webhook signature using the configured secret.
func (s *StripeGateway) ValidateWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	if s.webhookSecret == "" {
		return false, fmt.Errorf("webhook secret not configured")
	}

	parts := strings.Split(signature, ",")
	var timestamp, v1Signature string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, "t="); ok {
			timestamp = after
		} else if after, ok := strings.CutPrefix(part, "v1="); ok {
			v1Signature = after
		}
	}

	if timestamp == "" || v1Signature == "" {
		return false, fmt.Errorf("invalid signature format")
	}

	expectedSignature := s.computeHMACSHA256(timestamp, string(body), s.webhookSecret)
	return hmac.Equal([]byte(v1Signature), []byte(expectedSignature)), nil
}

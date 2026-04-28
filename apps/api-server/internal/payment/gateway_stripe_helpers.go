package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// makeStripeRequest issues an authenticated HTTP request to the Stripe API and
// decodes the JSON response (or returns a structured Stripe error).
func (s *StripeGateway) makeStripeRequest(ctx context.Context, method, path string, data any) (any, error) {
	url := "https://api.stripe.com" + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Version", "2023-10-16")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp map[string]any
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(map[string]any); ok {
				if message, ok := errorMsg["message"].(string); ok {
					return nil, fmt.Errorf("stripe error: %s", message)
				}
			}
		}
		return nil, fmt.Errorf("stripe error: %s", string(respBody))
	}

	var result any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// mapStripeStatus converts a Stripe-specific status into a PaymentStatusType.
func (s *StripeGateway) mapStripeStatus(status string) PaymentStatusType {
	switch status {
	case "requires_payment_method", "requires_confirmation", "requires_action":
		return PaymentStatusPending
	case "processing":
		return PaymentStatusProcessing
	case "succeeded":
		return PaymentStatusCompleted
	case "canceled":
		return PaymentStatusCancelled
	default:
		return PaymentStatusFailed
	}
}

// mapPaymentMethodType converts a PaymentMethodType to Stripe's string identifier.
func (s *StripeGateway) mapPaymentMethodType(pmt PaymentMethodType) string {
	switch pmt {
	case PaymentMethodTypeCreditCard:
		return "card"
	case PaymentMethodTypeBankAccount:
		return "bank_account"
	default:
		return "card"
	}
}

// computeHMACSHA256 returns the hex-encoded HMAC-SHA256 for timestamped payloads.
func (s *StripeGateway) computeHMACSHA256(timestamp, payload, secret string) string {
	data := timestamp + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

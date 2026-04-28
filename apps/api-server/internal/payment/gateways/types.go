package gateways

import (
	"encoding/json"
	"net/http"
	"time"
)

// StripeGateway implements payment processing via Stripe
type StripeGateway struct {
	secretKey     string
	webhookSecret string
	client        *http.Client
}

type PaymentRequest struct {
	SubscriberID uint    `json:"subscriber_id"`
	InvoiceID    uint    `json:"invoice_id"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	CustomerID   string  `json:"customer_id,omitempty"`
	Description  string  `json:"description,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	ClientSecret  string    `json:"client_secret"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

// CustomerResponse represents a customer creation response
type CustomerResponse struct {
	CustomerID string    `json:"customer_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID  string    `json:"refund_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentStatusResponse represents payment status response
type PaymentStatusResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	Type    string          `json:"type"`
	EventID string          `json:"event_id"`
	Created time.Time       `json:"created"`
	Data    json.RawMessage `json:"data"`
}

// InvoiceItem represents an invoice line item
type InvoiceItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Quantity    int     `json:"quantity"`
}

// InvoiceResponse represents an invoice response
type InvoiceResponse struct {
	InvoiceID string    `json:"invoice_id"`
	Status    string    `json:"status"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

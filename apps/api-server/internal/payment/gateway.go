package payment

import (
"context"
"time"
)

// PaymentGateway interface
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResponse, error)
	GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error)
	CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, paymentMethodID string) error
	ValidateWebhook(ctx context.Context, signature string, body []byte) (bool, error)
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Description     string            `json:"description"`
	CustomerID      string            `json:"customerId"`
	PaymentMethodID string            `json:"paymentMethodId"`
	Metadata        map[string]string `json:"metadata"`
	WebhookURL      string            `json:"webhookUrl"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	TransactionID string            `json:"transactionId"`
	Status        PaymentStatusType `json:"status"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	ProcessedAt   time.Time         `json:"processedAt"`
	FailureReason *string           `json:"failureReason,omitempty"`
	Metadata      map[string]string `json:"metadata"`
	RedirectURL   *string           `json:"redirectUrl,omitempty"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID      string            `json:"refundId"`
	TransactionID string            `json:"transactionId"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Status        PaymentStatusType `json:"status"`
	ProcessedAt   time.Time         `json:"processedAt"`
	FailureReason *string           `json:"failureReason,omitempty"`
}

// PaymentStatus represents payment status
type PaymentStatus struct {
	TransactionID string            `json:"transactionId"`
	Status        PaymentStatusType `json:"status"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	FailureReason *string           `json:"failureReason,omitempty"`
	Refunds       []RefundInfo      `json:"refunds,omitempty"`
}

// RefundInfo represents refund information
type RefundInfo struct {
	RefundID  string    `json:"refundId"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreatePaymentMethodRequest represents a payment method creation request
type CreatePaymentMethodRequest struct {
	Type       PaymentMethodType `json:"type"`
	Token      string            `json:"token"`
	CustomerID string            `json:"customerId"`
	IsDefault  bool              `json:"isDefault"`
	Metadata   map[string]string `json:"metadata"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID          string            `json:"id"`
	Type        PaymentMethodType `json:"type"`
	CustomerID  string            `json:"customerId"`
	Last4       string            `json:"last4"`
	Brand       string            `json:"brand"`
	ExpiryMonth int               `json:"expiryMonth"`
	ExpiryYear  int               `json:"expiryYear"`
	IsDefault   bool              `json:"isDefault"`
	CreatedAt   time.Time         `json:"createdAt"`
	Metadata    map[string]string `json:"metadata"`
}

// PaymentMethodDetails represents payment method details retrieved from token
type PaymentMethodDetails struct {
	PaymentMethodID string    `json:"paymentMethodId"`
	Type            string    `json:"type"`
	Last4           string    `json:"last4"`
	Brand           string    `json:"brand"`
	ExpiryMonth     int       `json:"expiryMonth"`
	ExpiryYear      int       `json:"expiryYear"`
	Fingerprint     string    `json:"fingerprint"`
	CreatedAt       time.Time `json:"createdAt"`
}

// PaymentStatusType represents payment status
type PaymentStatusType string

const (
	PaymentStatusPending    PaymentStatusType = "PENDING"
	PaymentStatusProcessing PaymentStatusType = "PROCESSING"
	PaymentStatusCompleted  PaymentStatusType = "COMPLETED"
	PaymentStatusFailed     PaymentStatusType = "FAILED"
	PaymentStatusRefunded   PaymentStatusType = "REFUNDED"
	PaymentStatusCancelled  PaymentStatusType = "CANCELLED"
)

// PaymentMethodType represents payment method type
type PaymentMethodType string

const (
	PaymentMethodTypeCreditCard  PaymentMethodType = "CREDIT_CARD"
	PaymentMethodTypeBankAccount PaymentMethodType = "BANK_ACCOUNT"
)

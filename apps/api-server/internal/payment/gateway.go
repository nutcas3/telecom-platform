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
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
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

// StripeGateway implements PaymentGateway for Stripe
type StripeGateway struct {
	apiKey        string
	webhookSecret string
	client        *http.Client
	config        *config.PaymentConfig
}

// NewStripeGateway creates a new Stripe payment gateway
func NewStripeGateway(cfg *config.PaymentConfig) *StripeGateway {
	return &StripeGateway{
		apiKey:        cfg.StripeAPIKey,
		webhookSecret: cfg.StripeWebhookSecret,
		client:        &http.Client{Timeout: 30 * time.Second},
		config:        cfg,
	}
}

// ProcessPayment processes a payment via Stripe
func (s *StripeGateway) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Create payment intent
	intentData := map[string]any{
		"amount":              int64(req.Amount * 100), // Stripe uses cents
		"currency":            strings.ToLower(req.Currency),
		"description":         req.Description,
		"customer":            req.CustomerID,
		"payment_method":      req.PaymentMethodID,
		"confirmation_method": "manual",
		"confirm":             true,
		"metadata":            req.Metadata,
	}

	if req.WebhookURL != "" {
		intentData["transfer_data"] = map[string]string{
			"destination": req.WebhookURL,
		}
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
		if lastPaymentError, ok := intent["last_payment_error"].(map[string]any); ok {
			if message, ok := lastPaymentError["message"].(string); ok {
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

// RefundPayment refunds a payment via Stripe
func (s *StripeGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResponse, error) {
	refundData := map[string]any{
		"payment_intent": transactionID,
		"amount":         int64(amount * 100), // Stripe uses cents
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
		Currency:      "USD", // Default currency, should be configurable
		Status:        status,
		ProcessedAt:   time.Now(),
		FailureReason: failureReason,
	}, nil
}

// GetPaymentStatus gets payment status from Stripe
func (s *StripeGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	intentResp, err := s.makeStripeRequest(ctx, "GET", "/v1/payment_intents/"+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	intent, ok := intentResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid payment intent response")
	}

	// Get refunds for this payment intent
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
	if lastPaymentError, ok := intent["last_payment_error"].(map[string]any); ok {
		if message, ok := lastPaymentError["message"].(string); ok {
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

// CreatePaymentMethod creates a payment method in Stripe
func (s *StripeGateway) CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*PaymentMethod, error) {
	pmData := map[string]any{
		"type": s.mapPaymentMethodType(req.Type),
		"card": map[string]string{
			"token": req.Token,
		},
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

	// Attach payment method to customer
	attachData := map[string]any{
		"customer": req.CustomerID,
	}
	_, err = s.makeStripeRequest(ctx, "POST", "/v1/payment_methods/"+pm["id"].(string)+"/attach", attachData)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	// Set as default if requested
	if req.IsDefault {
		customerData := map[string]any{
			"invoice_settings": map[string]string{
				"default_payment_method": pm["id"].(string),
			},
		}
		_, err = s.makeStripeRequest(ctx, "POST", "/v1/customers/"+req.CustomerID, customerData)
		if err != nil {
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

// DeletePaymentMethod deletes a payment method from Stripe
func (s *StripeGateway) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := s.makeStripeRequest(ctx, "DELETE", "/v1/payment_methods/"+paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}
	return nil
}

// ValidateWebhook validates a Stripe webhook
func (s *StripeGateway) ValidateWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	if s.webhookSecret == "" {
		return false, fmt.Errorf("webhook secret not configured")
	}

	// Parse signature
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

	// Create expected signature
	expectedSignature := s.computeHMACSHA256(timestamp, string(body), s.webhookSecret)

	return hmac.Equal([]byte(v1Signature), []byte(expectedSignature)), nil
}

// Helper methods
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

func (s *StripeGateway) computeHMACSHA256(timestamp, payload, secret string) string {
	data := timestamp + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// PaymentService handles payment operations
type PaymentService struct {
	gateway PaymentGateway
	db      *database.Database
	config  *config.PaymentConfig
}

// NewPaymentService creates a new payment service
func NewPaymentService(db *database.Database, cfg *config.PaymentConfig) *PaymentService {
	var gateway PaymentGateway

	switch cfg.Provider {
	case "stripe":
		gateway = NewStripeGateway(cfg)
	default:
		gateway = NewStripeGateway(cfg) // Default to Stripe
	}

	return &PaymentService{
		gateway: gateway,
		db:      db,
		config:  cfg,
	}
}

// ProcessTopUp processes a top-up payment
func (ps *PaymentService) ProcessTopUp(ctx context.Context, subscriberID int, amount float64, paymentMethodID string) (*models.Transaction, error) {
	// Get subscriber
	subscriber, err := ps.db.GetSubscriber(ctx, uint(subscriberID))
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Create payment request
	req := &PaymentRequest{
		Amount:          amount,
		Currency:        "USD",
		Description:     fmt.Sprintf("Top-up for subscriber %s", subscriber.MSISDN),
		CustomerID:      subscriber.MSISDN,
		PaymentMethodID: paymentMethodID,
		Metadata: map[string]string{
			"subscriber_id": fmt.Sprintf("%d", subscriberID),
			"type":          "top_up",
		},
	}

	// Process payment
	resp, err := ps.gateway.ProcessPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	// Create transaction record
	transaction := &models.Transaction{
		SubscriberID:  uint(subscriberID),
		TransactionID: resp.TransactionID,
		Type:          "TOP_UP",
		Amount:        amount,
		Currency:      resp.Currency,
		Status:        string(resp.Status),
		Description:   req.Description,
		CreatedAt:     resp.ProcessedAt,
		UpdatedAt:     resp.ProcessedAt,
	}

	// Save transaction
	err = ps.db.CreateTransaction(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// If payment was successful, update subscriber balance
	if resp.Status == PaymentStatusCompleted {
		err = ps.db.UpdateSubscriberBalance(ctx, uint(subscriberID), amount)
		if err != nil {
			return nil, fmt.Errorf("failed to update subscriber balance: %w", err)
		}
	}

	return transaction, nil
}

// GetPaymentStatus gets payment status
func (ps *PaymentService) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	return ps.gateway.GetPaymentStatus(ctx, transactionID)
}

// CreatePaymentMethod creates a payment method
func (ps *PaymentService) CreatePaymentMethod(ctx context.Context, subscriberID int, req *CreatePaymentMethodRequest) (*PaymentMethod, error) {
	subscriber, err := ps.db.GetSubscriber(ctx, uint(subscriberID))
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	req.CustomerID = subscriber.MSISDN

	pm, err := ps.gateway.CreatePaymentMethod(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	// Save payment method to database
	dbPM := &models.PaymentMethod{
		SubscriberID: uint(subscriberID),
		GatewayID:    pm.ID,
		Type:         models.PaymentMethodType(pm.Type),
		Last4:        pm.Last4,
		Brand:        pm.Brand,
		ExpiryMonth:  pm.ExpiryMonth,
		ExpiryYear:   pm.ExpiryYear,
		IsDefault:    pm.IsDefault,
		CreatedAt:    pm.CreatedAt,
	}

	err = ps.db.CreatePaymentMethod(ctx, dbPM)
	if err != nil {
		return nil, fmt.Errorf("failed to save payment method: %w", err)
	}

	pm.ID = fmt.Sprintf("%s", dbPM.ID) // Return database ID
	return pm, nil
}

// ValidateWebhook validates a webhook
func (ps *PaymentService) ValidateWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	return ps.gateway.ValidateWebhook(ctx, signature, body)
}

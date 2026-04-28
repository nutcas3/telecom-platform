package payment

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// PaymentService orchestrates payment flows using a configured PaymentGateway.
type PaymentService struct {
	gateway PaymentGateway
	db      *database.Database
	config  *config.PaymentConfig
}

// NewPaymentService creates a new payment service. Defaults to Stripe gateway.
func NewPaymentService(db *database.Database, cfg *config.PaymentConfig) *PaymentService {
	var gateway PaymentGateway
	switch cfg.Provider {
	case "stripe":
		gateway = NewStripeGateway(cfg)
	default:
		gateway = NewStripeGateway(cfg)
	}
	return &PaymentService{gateway: gateway, db: db, config: cfg}
}

// ProcessTopUp processes a top-up payment and records the transaction.
func (ps *PaymentService) ProcessTopUp(ctx context.Context, subscriberID int, amount float64, paymentMethodID string) (*models.Transaction, error) {
	subscriber, err := ps.db.GetSubscriber(ctx, uint(subscriberID))
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

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

	resp, err := ps.gateway.ProcessPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

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

	if err := ps.db.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	if resp.Status == PaymentStatusCompleted {
		if err := ps.db.UpdateSubscriberBalance(ctx, uint(subscriberID), amount); err != nil {
			return nil, fmt.Errorf("failed to update subscriber balance: %w", err)
		}
	}

	return transaction, nil
}

// GetPaymentStatus returns a payment status by transaction ID.
func (ps *PaymentService) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	return ps.gateway.GetPaymentStatus(ctx, transactionID)
}

// CreatePaymentMethod creates a payment method on the gateway and persists it.
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

	if err := ps.db.CreatePaymentMethod(ctx, dbPM); err != nil {
		return nil, fmt.Errorf("failed to save payment method: %w", err)
	}

	pm.ID = fmt.Sprintf("%s", dbPM.ID)
	return pm, nil
}

// ValidateWebhook validates a webhook using the configured gateway.
func (ps *PaymentService) ValidateWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	return ps.gateway.ValidateWebhook(ctx, signature, body)
}

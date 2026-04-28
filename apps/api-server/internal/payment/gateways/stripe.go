package gateways

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/invoiceitem"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/refund"
	"github.com/stripe/stripe-go/v82/webhook"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/payment"
)

// NewStripeGateway creates a new Stripe payment gateway
func NewStripeGateway(secretKey, webhookSecret string) *StripeGateway {
	stripe.Key = secretKey

	return &StripeGateway{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		client:        &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessPayment processes a payment via Stripe
func (g *StripeGateway) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   new(int64(req.Amount * 100)), // Convert to cents
		Currency: stripe.String(string(req.Currency)),
		Metadata: map[string]string{
			"subscriber_id": fmt.Sprintf("%d", req.SubscriberID),
			"invoice_id":    fmt.Sprintf("%d", req.InvoiceID),
		},
	}

	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &PaymentResponse{
		TransactionID: pi.ID,
		Status:        string(pi.Status),
		ClientSecret:  pi.ClientSecret,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CreatedAt:     time.Unix(pi.Created, 0),
	}, nil
}

// RetrievePaymentMethodFromToken retrieves payment method details from a Stripe token
func (g *StripeGateway) RetrievePaymentMethodFromToken(ctx context.Context, token string) (*payment.PaymentMethodDetails, error) {
	// Create payment method from token
	params := &stripe.PaymentMethodParams{
		Type: stripe.String("card"),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(token),
		},
	}

	pm, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method from token: %w", err)
	}

	// Extract card details
	if pm.Card == nil {
		return nil, fmt.Errorf("no card details found in payment method")
	}

	return &payment.PaymentMethodDetails{
		PaymentMethodID: pm.ID,
		Type:            string(pm.Type),
		Last4:           pm.Card.Last4,
		Brand:           string(pm.Card.Brand),
		ExpiryMonth:     int(pm.Card.ExpMonth),
		ExpiryYear:      int(pm.Card.ExpYear),
		Fingerprint:     pm.Card.Fingerprint,
		CreatedAt:       time.Unix(pm.Created, 0),
	}, nil
}

// CreateCustomer creates a Stripe customer
func (g *StripeGateway) CreateCustomer(ctx context.Context, subscriber *models.Subscriber) (*CustomerResponse, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(subscriber.Email),
		Name:  stripe.String(fmt.Sprintf("%s %s", subscriber.FirstName, subscriber.LastName)),
		Metadata: map[string]string{
			"subscriber_id": fmt.Sprintf("%d", subscriber.ID),
			"imsi":          string(subscriber.IMSI),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &CustomerResponse{
		CustomerID: cust.ID,
		Email:      cust.Email,
		Name:       cust.Name,
		CreatedAt:  time.Unix(cust.Created, 0),
	}, nil
}

// ProcessRefund processes a refund via Stripe
func (g *StripeGateway) ProcessRefund(ctx context.Context, transactionID string, amount float64) (*RefundResponse, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(transactionID),
	}

	if amount > 0 {
		params.Amount = new(int64(amount * 100)) // Convert to cents
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	return &RefundResponse{
		RefundID:  refund.ID,
		Amount:    float64(refund.Amount) / 100.0,
		Status:    string(refund.Status),
		CreatedAt: time.Unix(refund.Created, 0),
	}, nil
}

// GetPaymentStatus retrieves payment status from Stripe
func (g *StripeGateway) GetPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatusResponse, error) {
	pi, err := paymentintent.Get(transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	return &PaymentStatusResponse{
		TransactionID: pi.ID,
		Status:        string(pi.Status),
		Amount:        float64(pi.Amount) / 100.0,
		Currency:      string(pi.Currency),
		CreatedAt:     time.Unix(pi.Created, 0),
	}, nil
}

// ValidateWebhook validates and processes Stripe webhooks
func (g *StripeGateway) ValidateWebhook(ctx context.Context, signatureHeader string, body []byte) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(body, signatureHeader, g.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook validation failed: %w", err)
	}

	return &WebhookEvent{
		Type:    string(event.Type),
		EventID: event.ID,
		Created: time.Unix(event.Created, 0),
		Data:    event.Data.Raw,
	}, nil
}

// CreateInvoice creates a Stripe invoice
func (g *StripeGateway) CreateInvoice(ctx context.Context, customerID string, items []InvoiceItem) (*InvoiceResponse, error) {
	// Create invoice items first
	for _, item := range items {
		itemParams := &stripe.InvoiceItemParams{
			Customer:    stripe.String(customerID),
			Amount:      new(int64(item.Amount * 100)),
			Currency:    stripe.String(string(item.Currency)),
			Description: stripe.String(item.Description),
			Quantity:    new(int64(item.Quantity)),
		}

		_, err := invoiceitem.New(itemParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	// Create the invoice
	params := &stripe.InvoiceParams{
		Customer:    stripe.String(customerID),
		AutoAdvance: new(true),
	}

	inv, err := invoice.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return &InvoiceResponse{
		InvoiceID: inv.ID,
		Status:    string(inv.Status),
		Amount:    float64(inv.Total) / 100.0,
		Currency:  string(inv.Currency),
		CreatedAt: time.Unix(inv.Created, 0),
	}, nil
}

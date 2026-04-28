package telecom

import (
	"context"
	"fmt"
)

// PaymentAPI handles payment-related API calls
type PaymentAPI struct {
	client *HTTPClient
}

// NewPaymentAPI creates a new PaymentAPI
func NewPaymentAPI(client *HTTPClient) *PaymentAPI {
	return &PaymentAPI{client: client}
}

// CreateTransaction creates a new payment transaction
func (p *PaymentAPI) CreateTransaction(ctx context.Context, req *CreatePaymentRequest) (*PaymentTransaction, error) {
	var transaction PaymentTransaction
	err := p.client.Post(ctx, "/v1/payments/transactions", req, &transaction)
	return &transaction, err
}

// GetTransaction retrieves a payment transaction by ID
func (p *PaymentAPI) GetTransaction(ctx context.Context, transactionID string) (*PaymentTransaction, error) {
	var transaction PaymentTransaction
	err := p.client.Get(ctx, fmt.Sprintf("/v1/payments/transactions/%s", transactionID), &transaction)
	return &transaction, err
}

// ListTransactions retrieves a list of payment transactions
func (p *PaymentAPI) ListTransactions(ctx context.Context, subscriberID int64, status string, page, pageSize int32) (*SubscriberList, error) {
	params := map[string]string{
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}

	if subscriberID > 0 {
		params["subscriber_id"] = fmt.Sprintf("%d", subscriberID)
	}
	if status != "" {
		params["status"] = status
	}

	var list SubscriberList
	err := p.client.Get(ctx, "/v1/payments/transactions", &list, params)
	return &list, err
}

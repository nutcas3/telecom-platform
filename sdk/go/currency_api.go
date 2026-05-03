package telecom

import (
	"context"
	"fmt"
)

// CurrencyAPI provides currency and billing operations
type CurrencyAPI struct {
	client *HTTPClient
}

// NewCurrencyAPI creates a new currency API client
func NewCurrencyAPI(client *HTTPClient) *CurrencyAPI {
	return &CurrencyAPI{client: client}
}

// ConvertRequest represents a currency conversion request
type ConvertRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

// ConvertResponse represents a currency conversion response
type ConvertResponse struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Converted float64 `json:"converted"`
	Rate      float64 `json:"rate"`
	Timestamp string  `json:"timestamp"`
}

// ExchangeRate represents an exchange rate
type ExchangeRate struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Rate      float64 `json:"rate"`
	Timestamp string  `json:"timestamp"`
}

// BillingTransaction represents a billing transaction
type BillingTransaction struct {
	ID          string  `json:"id"`
	ProfileID   string  `json:"profile_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

// BillingSummary represents a billing summary
type BillingSummary struct {
	ProfileID        string             `json:"profile_id"`
	Period           string             `json:"period"`
	TotalAmount      float64            `json:"total_amount"`
	Currency         string             `json:"currency"`
	TransactionCount int                `json:"transaction_count"`
	Breakdown        map[string]float64 `json:"breakdown"`
}

// Convert converts currency
func (c *CurrencyAPI) Convert(ctx context.Context, req *ConvertRequest) (*ConvertResponse, error) {
	var result ConvertResponse
	err := c.client.Post(ctx, "/api/v1/currency/convert", req, &result)
	return &result, err
}

// GetExchangeRate gets the exchange rate between currencies
func (c *CurrencyAPI) GetExchangeRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	var result ExchangeRate
	err := c.client.Get(ctx, fmt.Sprintf("/api/v1/currency/exchange/%s/%s", from, to), &result)
	return &result, err
}

// GetExchangeRateHistory gets exchange rate history
func (c *CurrencyAPI) GetExchangeRateHistory(ctx context.Context, from, to string, days int) ([]ExchangeRate, error) {
	var result []ExchangeRate
	err := c.client.Get(ctx, fmt.Sprintf("/api/v1/currency/exchange/%s/%s/history", from, to), &result, map[string]string{"days": fmt.Sprintf("%d", days)})
	return result, err
}

// GetSupportedCurrencies gets list of supported currencies
func (c *CurrencyAPI) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	var result []string
	err := c.client.Get(ctx, "/api/v1/currency/currencies", &result)
	return result, err
}

// RefreshExchangeRates refreshes exchange rates
func (c *CurrencyAPI) RefreshExchangeRates(ctx context.Context) error {
	return c.client.Post(ctx, "/api/v1/currency/exchange/refresh", nil, nil)
}

// ProcessBilling processes a billing transaction
func (c *CurrencyAPI) ProcessBilling(ctx context.Context, billingData map[string]interface{}) (*BillingTransaction, error) {
	var result BillingTransaction
	err := c.client.Post(ctx, "/api/v1/currency/billing", billingData, &result)
	return &result, err
}

// GetBillingHistory gets billing history for a profile
func (c *CurrencyAPI) GetBillingHistory(ctx context.Context, profileID string, limit int) ([]BillingTransaction, error) {
	var result []BillingTransaction
	err := c.client.Get(ctx, fmt.Sprintf("/api/v1/currency/billing/history/%s", profileID), &result, map[string]string{"limit": fmt.Sprintf("%d", limit)})
	return result, err
}

// GetBillingSummary gets billing summary for a profile
func (c *CurrencyAPI) GetBillingSummary(ctx context.Context, profileID, period string) (*BillingSummary, error) {
	var result BillingSummary
	err := c.client.Get(ctx, fmt.Sprintf("/api/v1/currency/billing/summary/%s", profileID), &result, map[string]string{"period": period})
	return &result, err
}

// ProcessRefund processes a refund
func (c *CurrencyAPI) ProcessRefund(ctx context.Context, transactionID, reason string) (*BillingTransaction, error) {
	var result BillingTransaction
	err := c.client.Post(ctx, fmt.Sprintf("/api/v1/currency/billing/refund/%s", transactionID), map[string]string{"reason": reason}, &result)
	return &result, err
}

// GetBillingAnalytics gets billing analytics
func (c *CurrencyAPI) GetBillingAnalytics(ctx context.Context, period string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.client.Get(ctx, "/api/v1/currency/billing/analytics", &result, map[string]string{"period": period})
	return result, err
}

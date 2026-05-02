package currency

import (
	"context"
	"time"
)

// Repository defines the interface for currency data operations
type Repository interface {
	// Currency operations
	CreateCurrency(ctx context.Context, currency *Currency) error
	GetCurrency(ctx context.Context, code string) (*Currency, error)
	UpdateCurrency(ctx context.Context, currency *Currency) error
	DeleteCurrency(ctx context.Context, code string) error
	ListCurrencies(ctx context.Context, filter *CurrencyFilter) ([]*Currency, error)
	CountCurrencies(ctx context.Context, filter *CurrencyFilter) (int, error)

	// Exchange rate operations
	CreateExchangeRate(ctx context.Context, rate *ExchangeRate) error
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error)
	UpdateExchangeRate(ctx context.Context, rate *ExchangeRate) error
	DeleteExchangeRate(ctx context.Context, id string) error
	ListExchangeRates(ctx context.Context, filter *ExchangeRateFilter) ([]*ExchangeRate, error)
	GetLatestExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, id string) error
	ListTransactions(ctx context.Context, filter *TransactionFilter) ([]*Transaction, error)
	CountTransactions(ctx context.Context, filter *TransactionFilter) (int, error)
}

// ExchangeRateProvider defines the interface for external exchange rate providers
type ExchangeRateProvider interface {
	GetRates(ctx context.Context) (map[string]float64, error)
	GetRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error)
	RefreshRates(ctx context.Context) error
}

// ExchangeRateService defines the interface for exchange rate operations
type ExchangeRateService interface {
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error)
	ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (*CurrencyConversionResponse, error)
	GetRateHistory(ctx context.Context, fromCurrency, toCurrency string, days int) ([]*ExchangeRate, error)
	UpdateExchangeRate(ctx context.Context, rate *ExchangeRate) error
	GetSupportedCurrencies(ctx context.Context) ([]*Currency, error)
	ValidateCurrencyPair(ctx context.Context, fromCurrency, toCurrency string) error
	RefreshRates(ctx context.Context) error
}

// BillingService defines the interface for multi-currency billing operations
type BillingService interface {
	ProcessBilling(ctx context.Context, req *BillingRequest) (*BillingResponse, error)
	ConvertAmount(ctx context.Context, req *CurrencyConversionRequest) (*CurrencyConversionResponse, error)
	GetBillingHistory(ctx context.Context, profileID string, filter *TransactionFilter) ([]*Transaction, error)
	CalculateTotalBilling(ctx context.Context, profileID string, fromDate, toDate time.Time) (*BillingSummary, error)
}

// AnalyticsService defines the interface for currency analytics
type AnalyticsService interface {
	GetRevenueByCurrency(ctx context.Context, filter *TransactionFilter) (map[string]float64, error)
	GetTransactionVolumeByCurrency(ctx context.Context, filter *TransactionFilter) (map[string]int64, error)
	GetExchangeRateTrends(ctx context.Context, fromCurrency, toCurrency string, days int) ([]*ExchangeRate, error)
	GetCurrencyUsageStats(ctx context.Context) (*CurrencyUsageStats, error)
}

// BillingSummary represents a billing summary for a profile
type BillingSummary struct {
	ProfileID        string         `json:"profile_id"`
	TotalAmount      float64        `json:"total_amount"`
	Currency         string         `json:"currency"`
	BaseTotalAmount  float64        `json:"base_total_amount"`
	BaseCurrency     string         `json:"base_currency"`
	TransactionCount int            `json:"transaction_count"`
	FromDate         time.Time      `json:"from_date"`
	ToDate           time.Time      `json:"to_date"`
	Breakdown        map[string]any `json:"breakdown"`
}

// CurrencyUsageStats represents statistics about currency usage
type CurrencyUsageStats struct {
	TotalCurrencies      int              `json:"total_currencies"`
	ActiveCurrencies     int              `json:"active_currencies"`
	TotalTransactions    int64            `json:"total_transactions"`
	TotalVolume          float64          `json:"total_volume"`
	MostUsedCurrency     string           `json:"most_used_currency"`
	CurrencyDistribution map[string]int64 `json:"currency_distribution"`
	ExchangeRateCount    int              `json:"exchange_rate_count"`
	LastUpdated          time.Time        `json:"last_updated"`
}

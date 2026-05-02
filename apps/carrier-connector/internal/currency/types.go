package currency

import (
	"time"
)

// Currency represents a currency with its properties
type Currency struct {
	Code             string    `json:"code" db:"code"`                           // ISO 4217 currency code (USD, EUR, etc.)
	Name             string    `json:"name" db:"name"`                           // Full currency name
	Symbol           string    `json:"symbol" db:"symbol"`                       // Currency symbol ($, €, etc.)
	DecimalPlaces    int       `json:"decimal_places" db:"decimal_places"`       // Number of decimal places
	IsActive         bool      `json:"is_active" db:"is_active"`                 // Whether currency is active
	SupportedRegions []string  `json:"supported_regions" db:"supported_regions"` // Supported regions/countries
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
	ID           string     `json:"id" db:"id"`
	FromCurrency string     `json:"from_currency" db:"from_currency"`
	ToCurrency   string     `json:"to_currency" db:"to_currency"`
	Rate         float64    `json:"rate" db:"rate"`                   // Exchange rate (1 FromCurrency = Rate ToCurrency)
	Source       string     `json:"source" db:"source"`               // Data source (ECB, FED, etc.)
	ValidFrom    time.Time  `json:"valid_from" db:"valid_from"`       // When rate becomes valid
	ValidTo      *time.Time `json:"valid_to,omitempty" db:"valid_to"` // When rate expires (optional)
	IsActive     bool       `json:"is_active" db:"is_active"`         // Whether rate is currently active
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Transaction represents a financial transaction in multi-currency context
type Transaction struct {
	ID             string            `json:"id" db:"id"`
	ProfileID      string            `json:"profile_id" db:"profile_id"`
	SubscriptionID string            `json:"subscription_id" db:"subscription_id"`
	Type           TransactionType   `json:"type" db:"type"`
	Amount         float64           `json:"amount" db:"amount"`
	Currency       string            `json:"currency" db:"currency"`
	BaseAmount     float64           `json:"base_amount" db:"base_amount"`     // Amount in base currency (USD)
	BaseCurrency   string            `json:"base_currency" db:"base_currency"` // Base currency for reporting
	ExchangeRate   float64           `json:"exchange_rate" db:"exchange_rate"` // Rate used for conversion
	Description    string            `json:"description" db:"description"`
	Status         TransactionStatus `json:"status" db:"status"`
	Metadata       map[string]any    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
}

// TransactionType defines the type of transaction
type TransactionType string

const (
	TransactionTypeSubscription TransactionType = "subscription"
	TransactionTypeUsage        TransactionType = "usage"
	TransactionTypeOverage      TransactionType = "overage"
	TransactionTypeRefund       TransactionType = "refund"
	TransactionTypeAdjustment   TransactionType = "adjustment"
	TransactionTypeDiscount     TransactionType = "discount"
)

// TransactionStatus defines the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// CurrencyFilter defines filtering options for currency queries
type CurrencyFilter struct {
	IsActive  *bool  `json:"is_active,omitempty"`
	Region    string `json:"region,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// ExchangeRateFilter defines filtering options for exchange rate queries
type ExchangeRateFilter struct {
	FromCurrency string `json:"from_currency,omitempty"`
	ToCurrency   string `json:"to_currency,omitempty"`
	Source       string `json:"source,omitempty"`
	IsValid      *bool  `json:"is_valid,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
	SortBy       string `json:"sort_by,omitempty"`
	SortOrder    string `json:"sort_order,omitempty"`
}

// TransactionFilter defines filtering options for transaction queries
type TransactionFilter struct {
	ProfileID      string            `json:"profile_id,omitempty"`
	SubscriptionID string            `json:"subscription_id,omitempty"`
	Type           TransactionType   `json:"type,omitempty"`
	Status         TransactionStatus `json:"status,omitempty"`
	Currency       string            `json:"currency,omitempty"`
	FromDate       *time.Time        `json:"from_date,omitempty"`
	ToDate         *time.Time        `json:"to_date,omitempty"`
	MinAmount      float64           `json:"min_amount,omitempty"`
	MaxAmount      float64           `json:"max_amount,omitempty"`
	Limit          int               `json:"limit,omitempty"`
	Offset         int               `json:"offset,omitempty"`
	SortBy         string            `json:"sort_by,omitempty"`
	SortOrder      string            `json:"sort_order,omitempty"`
}

// CurrencyConversionRequest represents a request to convert an amount between currencies
type CurrencyConversionRequest struct {
	Amount       float64 `json:"amount" binding:"required,min=0"`
	FromCurrency string  `json:"from_currency" binding:"required"`
	ToCurrency   string  `json:"to_currency" binding:"required"`
}

// CurrencyConversionResponse represents the response from currency conversion
type CurrencyConversionResponse struct {
	OriginalAmount    float64   `json:"original_amount"`
	OriginalCurrency  string    `json:"original_currency"`
	ConvertedAmount   float64   `json:"converted_amount"`
	ConvertedCurrency string    `json:"converted_currency"`
	ExchangeRate      float64   `json:"exchange_rate"`
	ConvertedAt       time.Time `json:"converted_at"`
}

// BillingRequest represents a billing request in multi-currency context
type BillingRequest struct {
	ProfileID      string    `json:"profile_id" binding:"required"`
	SubscriptionID string    `json:"subscription_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,min=0"`
	Currency       string    `json:"currency" binding:"required"`
	Description    string    `json:"description"`
	BillingDate    time.Time `json:"billing_date"`
}

// BillingResponse represents the response from billing operation
type BillingResponse struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	BaseAmount    float64   `json:"base_amount"`
	BaseCurrency  string    `json:"base_currency"`
	ExchangeRate  float64   `json:"exchange_rate"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// CurrencyRevenue represents revenue for a specific currency
type CurrencyRevenue struct {
	Currency string  `json:"currency"`
	Revenue  float64 `json:"revenue"`
}

// TransactionTypeStats represents statistics for a transaction type
type TransactionTypeStats struct {
	Type     TransactionType `json:"type"`
	Count    int             `json:"count"`
	Amount   float64         `json:"amount"`
	Currency string          `json:"currency"`
}

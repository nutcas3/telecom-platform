package currency

import (
	"time"

	"gorm.io/gorm"
)

// CurrencyModel represents the database model for currencies
type CurrencyModel struct {
	Code             string    `gorm:"primaryKey;column:code" json:"code"`
	Name             string    `gorm:"column:name" json:"name"`
	Symbol           string    `gorm:"column:symbol" json:"symbol"`
	DecimalPlaces    int       `gorm:"column:decimal_places" json:"decimal_places"`
	IsActive         bool      `gorm:"column:is_active" json:"is_active"`
	SupportedRegions string    `gorm:"column:supported_regions;type:text" json:"supported_regions"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the currency model
func (CurrencyModel) TableName() string {
	return "currencies"
}

// ExchangeRateModel represents the database model for exchange rates
type ExchangeRateModel struct {
	ID           string     `gorm:"primaryKey;column:id" json:"id"`
	FromCurrency string     `gorm:"column:from_currency;index" json:"from_currency"`
	ToCurrency   string     `gorm:"column:to_currency;index" json:"to_currency"`
	Rate         float64    `gorm:"column:rate" json:"rate"`
	Source       string     `gorm:"column:source" json:"source"`
	ValidFrom    time.Time  `gorm:"column:valid_from;index" json:"valid_from"`
	ValidTo      *time.Time `gorm:"column:valid_to;index" json:"valid_to"`
	IsActive     bool       `gorm:"column:is_active" json:"is_active"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the exchange rate model
func (ExchangeRateModel) TableName() string {
	return "exchange_rates"
}

// TransactionModel represents the database model for transactions
type TransactionModel struct {
	ID             string    `gorm:"primaryKey;column:id" json:"id"`
	ProfileID      string    `gorm:"column:profile_id;index" json:"profile_id"`
	SubscriptionID string    `gorm:"column:subscription_id;index" json:"subscription_id"`
	Type           string    `gorm:"column:type;index" json:"type"`
	Amount         float64   `gorm:"column:amount" json:"amount"`
	Currency       string    `gorm:"column:currency;index" json:"currency"`
	BaseAmount     float64   `gorm:"column:base_amount" json:"base_amount"`
	BaseCurrency   string    `gorm:"column:base_currency" json:"base_currency"`
	ExchangeRate   float64   `gorm:"column:exchange_rate" json:"exchange_rate"`
	Description    string    `gorm:"column:description" json:"description"`
	Status         string    `gorm:"column:status;index" json:"status"`
	Metadata       string    `gorm:"column:metadata;type:text" json:"metadata"`
	CreatedAt      time.Time `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the transaction model
func (TransactionModel) TableName() string {
	return "transactions"
}

// BeforeCreate hook for CurrencyModel
func (c *CurrencyModel) BeforeCreate(tx *gorm.DB) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for CurrencyModel
func (c *CurrencyModel) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for ExchangeRateModel
func (e *ExchangeRateModel) BeforeCreate(tx *gorm.DB) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	e.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for ExchangeRateModel
func (e *ExchangeRateModel) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for TransactionModel
func (t *TransactionModel) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for TransactionModel
func (t *TransactionModel) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

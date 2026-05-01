package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
)

// currencyToModel converts currency domain model to database model
func (r *GormRepository) currencyToModel(currency *currency.Currency) *currency.CurrencyModel {
	return &currency.CurrencyModel{
		Code:             currency.Code,
		Name:             currency.Name,
		Symbol:           currency.Symbol,
		DecimalPlaces:    currency.DecimalPlaces,
		IsActive:         currency.IsActive,
		SupportedRegions: strings.Join(currency.SupportedRegions, ","),
		CreatedAt:        currency.CreatedAt,
		UpdatedAt:        currency.UpdatedAt,
	}
}

// modelToCurrency converts database model to currency domain model
func (r *GormRepository) modelToCurrency(model *currency.CurrencyModel) (*currency.Currency, error) {
	var regions []string
	if model.SupportedRegions != "" {
		regions = strings.Split(model.SupportedRegions, ",")
	}

	return &currency.Currency{
		Code:             model.Code,
		Name:             model.Name,
		Symbol:           model.Symbol,
		DecimalPlaces:    model.DecimalPlaces,
		IsActive:         model.IsActive,
		SupportedRegions: regions,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}, nil
}

// exchangeRateToModel converts exchange rate domain model to database model
func (r *GormRepository) exchangeRateToModel(rate *currency.ExchangeRate) *currency.ExchangeRateModel {
	return &currency.ExchangeRateModel{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Source:       rate.Source,
		ValidFrom:    rate.ValidFrom,
		ValidTo:      rate.ValidTo,
		IsActive:     rate.IsActive,
		CreatedAt:    rate.CreatedAt,
		UpdatedAt:    rate.UpdatedAt,
	}
}

// modelToExchangeRate converts database model to exchange rate domain model
func (r *GormRepository) modelToExchangeRate(model *currency.ExchangeRateModel) (*currency.ExchangeRate, error) {
	return &currency.ExchangeRate{
		ID:           model.ID,
		FromCurrency: model.FromCurrency,
		ToCurrency:   model.ToCurrency,
		Rate:         model.Rate,
		Source:       model.Source,
		ValidFrom:    model.ValidFrom,
		ValidTo:      model.ValidTo,
		IsActive:     model.IsActive,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// transactionToModel converts transaction domain model to database model
func (r *GormRepository) transactionToModel(transaction *currency.Transaction) *currency.TransactionModel {
	metadata := ""
	if transaction.Metadata != nil {
		if data, err := json.Marshal(transaction.Metadata); err == nil {
			metadata = string(data)
		}
	}

	return &currency.TransactionModel{
		ID:             transaction.ID,
		ProfileID:      transaction.ProfileID,
		SubscriptionID: transaction.SubscriptionID,
		Type:           string(transaction.Type),
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		BaseAmount:     transaction.BaseAmount,
		BaseCurrency:   transaction.BaseCurrency,
		ExchangeRate:   transaction.ExchangeRate,
		Description:    transaction.Description,
		Status:         string(transaction.Status),
		Metadata:       metadata,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}
}

// modelToTransaction converts database model to transaction domain model
func (r *GormRepository) modelToTransaction(model *currency.TransactionModel) (*currency.Transaction, error) {
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	return &currency.Transaction{
		ID:             model.ID,
		ProfileID:      model.ProfileID,
		SubscriptionID: model.SubscriptionID,
		Type:           currency.TransactionType(model.Type),
		Amount:         model.Amount,
		Currency:       model.Currency,
		BaseAmount:     model.BaseAmount,
		BaseCurrency:   model.BaseCurrency,
		ExchangeRate:   model.ExchangeRate,
		Description:    model.Description,
		Status:         currency.TransactionStatus(model.Status),
		Metadata:       metadata,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

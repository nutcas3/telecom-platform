package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
)

// ListExchangeRates retrieves exchange rates based on filter
func (r *GormRepository) ListExchangeRates(ctx context.Context, filter *currency.ExchangeRateFilter) ([]*currency.ExchangeRate, error) {
	query := r.db.WithContext(ctx).Model(&currency.ExchangeRateModel{})

	// Apply filters
	if filter.FromCurrency != "" {
		query = query.Where("from_currency = ?", filter.FromCurrency)
	}
	if filter.ToCurrency != "" {
		query = query.Where("to_currency = ?", filter.ToCurrency)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.IsValid != nil {
		now := time.Now()
		if *filter.IsValid {
			query = query.Where("valid_from <= ? AND (valid_to IS NULL OR valid_to >= ?)", now, now)
		} else {
			query = query.Where("valid_to < ?", now)
		}
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("from_currency, to_currency, valid_from DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var models []currency.ExchangeRateModel
	if err := query.Find(&models).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list exchange rates")
		return nil, fmt.Errorf("failed to list exchange rates: %w", err)
	}

	rates := make([]*currency.ExchangeRate, 0, len(models))
	for _, model := range models {
		rate, err := r.modelToExchangeRate(&model)
		if err != nil {
			r.logger.WithError(err).Error("Failed to convert exchange rate model")
			continue
		}
		rates = append(rates, rate)
	}

	return rates, nil
}

// GetLatestExchangeRate gets the latest valid exchange rate
func (r *GormRepository) GetLatestExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*currency.ExchangeRate, error) {
	now := time.Now()
	var model currency.ExchangeRateModel

	if err := r.db.WithContext(ctx).
		Where("from_currency = ? AND to_currency = ? AND is_active = ?", fromCurrency, toCurrency, true).
		Where("valid_from <= ? AND (valid_to IS NULL OR valid_to >= ?)", now, now).
		Order("valid_from DESC").
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no valid exchange rate found: %s to %s", fromCurrency, toCurrency)
		}
		r.logger.WithError(err).Error("Failed to get latest exchange rate")
		return nil, fmt.Errorf("failed to get latest exchange rate: %w", err)
	}

	return r.modelToExchangeRate(&model)
}

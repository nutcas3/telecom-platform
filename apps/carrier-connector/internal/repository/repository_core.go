package repository

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GormRepository implements the currency repository using GORM
type GormRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewGormRepository creates a new GORM currency repository
func NewGormRepository(db *gorm.DB, logger *logrus.Logger) *GormRepository {
	return &GormRepository{
		db:     db,
		logger: logger,
	}
}

// CreateCurrency creates a new currency
func (r *GormRepository) CreateCurrency(ctx context.Context, currency *currency.Currency) error {
	model := r.currencyToModel(currency)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create currency")
		return fmt.Errorf("failed to create currency: %w", err)
	}

	return nil
}

// GetCurrency retrieves a currency by code
func (r *GormRepository) GetCurrency(ctx context.Context, code string) (*currency.Currency, error) {
	var model currency.CurrencyModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("currency not found: %s", code)
		}
		r.logger.WithError(err).Error("Failed to get currency")
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	return r.modelToCurrency(&model)
}

// UpdateCurrency updates an existing currency
func (r *GormRepository) UpdateCurrency(ctx context.Context, currency *currency.Currency) error {
	model := r.currencyToModel(currency)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update currency")
		return fmt.Errorf("failed to update currency: %w", err)
	}

	return nil
}

// DeleteCurrency deletes a currency
func (r *GormRepository) DeleteCurrency(ctx context.Context, code string) error {
	if err := r.db.WithContext(ctx).Delete(&currency.CurrencyModel{}, "code = ?", code).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete currency")
		return fmt.Errorf("failed to delete currency: %w", err)
	}

	return nil
}

// ListCurrencies retrieves currencies based on filter
func (r *GormRepository) ListCurrencies(ctx context.Context, filter *currency.CurrencyFilter) ([]*currency.Currency, error) {
	query := r.db.WithContext(ctx).Model(&currency.CurrencyModel{})

	// Apply filters
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Region != "" {
		query = query.Where("supported_regions LIKE ?", "%"+filter.Region+"%")
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("code ASC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var models []currency.CurrencyModel
	if err := query.Find(&models).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list currencies")
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}

	currencies := make([]*currency.Currency, 0, len(models))
	for _, model := range models {
		curr, err := r.modelToCurrency(&model)
		if err != nil {
			r.logger.WithError(err).Error("Failed to convert currency model")
			continue
		}
		currencies = append(currencies, curr)
	}

	return currencies, nil
}

// CountCurrencies counts currencies based on filter
func (r *GormRepository) CountCurrencies(ctx context.Context, filter *currency.CurrencyFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&currency.CurrencyModel{})

	// Apply filters
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Region != "" {
		query = query.Where("supported_regions LIKE ?", "%"+filter.Region+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count currencies")
		return 0, fmt.Errorf("failed to count currencies: %w", err)
	}

	return int(count), nil
}

// CreateExchangeRate creates a new exchange rate
func (r *GormRepository) CreateExchangeRate(ctx context.Context, rate *currency.ExchangeRate) error {
	model := r.exchangeRateToModel(rate)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create exchange rate")
		return fmt.Errorf("failed to create exchange rate: %w", err)
	}

	return nil
}

// GetExchangeRate retrieves an exchange rate by ID
func (r *GormRepository) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*currency.ExchangeRate, error) {
	var model currency.ExchangeRateModel
	if err := r.db.WithContext(ctx).Where("from_currency = ? AND to_currency = ?", fromCurrency, toCurrency).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("exchange rate not found: %s to %s", fromCurrency, toCurrency)
		}
		r.logger.WithError(err).Error("Failed to get exchange rate")
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	return r.modelToExchangeRate(&model)
}

// UpdateExchangeRate updates an existing exchange rate
func (r *GormRepository) UpdateExchangeRate(ctx context.Context, rate *currency.ExchangeRate) error {
	model := r.exchangeRateToModel(rate)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update exchange rate")
		return fmt.Errorf("failed to update exchange rate: %w", err)
	}

	return nil
}

// DeleteExchangeRate deletes an exchange rate
func (r *GormRepository) DeleteExchangeRate(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&currency.ExchangeRateModel{}, "id = ?", id).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete exchange rate")
		return fmt.Errorf("failed to delete exchange rate: %w", err)
	}

	return nil
}

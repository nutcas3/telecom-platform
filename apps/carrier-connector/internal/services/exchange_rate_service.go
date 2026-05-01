package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/sirupsen/logrus"
)

type ExchangeRateService struct {
	repository   currency.Repository
	logger       *logrus.Logger
	providers    []currency.ExchangeRateProvider
	baseCurrency string
}

func NewExchangeRateService(repository currency.Repository, logger *logrus.Logger, baseCurrency string) *ExchangeRateService {
	return &ExchangeRateService{
		repository:   repository,
		logger:       logger,
		providers:    make([]currency.ExchangeRateProvider, 0),
		baseCurrency: baseCurrency,
	}
}

func (s *ExchangeRateService) AddProvider(provider currency.ExchangeRateProvider) {
	s.providers = append(s.providers, provider)
}

func (s *ExchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*currency.ExchangeRate, error) {
	if fromCurrency == toCurrency {
		return &currency.ExchangeRate{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         1.0,
			Source:       "direct",
			ValidFrom:    time.Now(),
			IsActive:     true,
		}, nil
	}

	rate, err := s.repository.GetLatestExchangeRate(ctx, fromCurrency, toCurrency)
	if err == nil {
		return rate, nil
	}

	for _, provider := range s.providers {
		providerRate, err := provider.GetRate(ctx, fromCurrency, toCurrency)
		if err == nil {
			newRate := &currency.ExchangeRate{
				ID:           fmt.Sprintf("%s_%s_%d", fromCurrency, toCurrency, time.Now().Unix()),
				FromCurrency: fromCurrency,
				ToCurrency:   toCurrency,
				Rate:         providerRate,
				Source:       "provider",
				ValidFrom:    time.Now(),
				IsActive:     true,
			}

			if err := s.repository.CreateExchangeRate(ctx, newRate); err != nil {
				s.logger.WithError(err).Error("Failed to save exchange rate")
			}

			return newRate, nil
		}
	}

	return nil, fmt.Errorf("exchange rate not found: %s to %s", fromCurrency, toCurrency)
}

func (s *ExchangeRateService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (*currency.CurrencyConversionResponse, error) {
	rate, err := s.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	convertedAmount := amount * rate.Rate

	return &currency.CurrencyConversionResponse{
		OriginalAmount:    amount,
		OriginalCurrency:  fromCurrency,
		ConvertedAmount:   convertedAmount,
		ConvertedCurrency: toCurrency,
		ExchangeRate:      rate.Rate,
		ConvertedAt:       time.Now(),
	}, nil
}

func (s *ExchangeRateService) RefreshRates(ctx context.Context) error {
	s.logger.Info("Refreshing exchange rates")

	for _, provider := range s.providers {
		if err := provider.RefreshRates(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to refresh rates from provider")
			continue
		}
	}

	s.logger.Info("Exchange rates refreshed successfully")
	return nil
}

func (s *ExchangeRateService) GetRateHistory(ctx context.Context, fromCurrency, toCurrency string, days int) ([]*currency.ExchangeRate, error) {
	filter := &currency.ExchangeRateFilter{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		IsValid:      &[]bool{false}[0], // Include historical rates
		Limit:        days,
		SortBy:       "valid_from",
		SortOrder:    "desc",
	}

	rates, err := s.repository.ListExchangeRates(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate history: %w", err)
	}

	return rates, nil
}

func (s *ExchangeRateService) UpdateExchangeRate(ctx context.Context, rate *currency.ExchangeRate) error {
	if rate.Rate <= 0 {
		return fmt.Errorf("invalid exchange rate: must be positive")
	}

	if rate.FromCurrency == rate.ToCurrency {
		return fmt.Errorf("invalid currency pair: from and to currencies cannot be the same")
	}

	now := time.Now()
	rate.ValidFrom = now
	rate.IsActive = true

	filter := &currency.ExchangeRateFilter{
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		IsValid:      &[]bool{true}[0],
	}

	oldRates, err := s.repository.ListExchangeRates(ctx, filter)
	if err == nil {
		for _, oldRate := range oldRates {
			oldRate.IsActive = false
			if err := s.repository.UpdateExchangeRate(ctx, oldRate); err != nil {
				s.logger.WithError(err).Error("Failed to deactivate old exchange rate")
			}
		}
	}

	rate.ID = fmt.Sprintf("%s_%s_%d", rate.FromCurrency, rate.ToCurrency, time.Now().Unix())
	if err := s.repository.CreateExchangeRate(ctx, rate); err != nil {
		return fmt.Errorf("failed to create exchange rate: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"from_currency": rate.FromCurrency,
		"to_currency":   rate.ToCurrency,
		"rate":          rate.Rate,
		"source":        rate.Source,
	}).Info("Exchange rate updated")

	return nil
}

func (s *ExchangeRateService) GetSupportedCurrencies(ctx context.Context) ([]*currency.Currency, error) {
	filter := &currency.CurrencyFilter{
		IsActive: &[]bool{true}[0],
	}

	currencies, err := s.repository.ListCurrencies(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported currencies: %w", err)
	}

	return currencies, nil
}

func (s *ExchangeRateService) ValidateCurrencyPair(ctx context.Context, fromCurrency, toCurrency string) error {
	_, err := s.repository.GetCurrency(ctx, fromCurrency)
	if err != nil {
		return fmt.Errorf("unsupported from currency: %s", fromCurrency)
	}

	_, err = s.repository.GetCurrency(ctx, toCurrency)
	if err != nil {
		return fmt.Errorf("unsupported to currency: %s", toCurrency)
	}

	_, err = s.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return fmt.Errorf("no exchange rate available: %s to %s", fromCurrency, toCurrency)
	}

	return nil
}

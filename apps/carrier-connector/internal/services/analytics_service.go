package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/sirupsen/logrus"
)

// AnalyticsServiceImpl handles currency analytics operations
type AnalyticsServiceImpl struct {
	repository currency.Repository
	logger     *logrus.Logger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repository currency.Repository, logger *logrus.Logger) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		repository: repository,
		logger:     logger,
	}
}

// GetRevenueByCurrency calculates revenue breakdown by currency
func (s *AnalyticsServiceImpl) GetRevenueByCurrency(ctx context.Context, filter *currency.TransactionFilter) (map[string]float64, error) {
	if filter == nil {
		filter = &currency.TransactionFilter{}
	}

	filter.Status = currency.TransactionStatusCompleted

	transactions, err := s.repository.ListTransactions(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get transactions for revenue analysis")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	revenueByCurrency := make(map[string]float64)

	for _, tx := range transactions {
		if tx.Type == currency.TransactionTypeSubscription || tx.Type == currency.TransactionTypeUsage || tx.Type == currency.TransactionTypeOverage {
			revenueByCurrency[tx.Currency] += tx.Amount
		}
	}

	return revenueByCurrency, nil
}

// GetTransactionVolumeByCurrency calculates transaction volume by currency
func (s *AnalyticsServiceImpl) GetTransactionVolumeByCurrency(ctx context.Context, filter *currency.TransactionFilter) (map[string]int64, error) {
	if filter == nil {
		filter = &currency.TransactionFilter{}
	}

	transactions, err := s.repository.ListTransactions(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get transactions for volume analysis")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	volumeByCurrency := make(map[string]int64)

	for _, tx := range transactions {
		volumeByCurrency[tx.Currency]++
	}

	return volumeByCurrency, nil
}

// GetExchangeRateTrends retrieves exchange rate trends for a currency pair
func (s *AnalyticsServiceImpl) GetExchangeRateTrends(ctx context.Context, fromCurrency, toCurrency string, days int) ([]*currency.ExchangeRate, error) {
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
		s.logger.WithError(err).Error("Failed to get exchange rate trends")
		return nil, fmt.Errorf("failed to get exchange rate trends: %w", err)
	}

	return rates, nil
}

// GetCurrencyUsageStats retrieves currency usage statistics
func (s *AnalyticsServiceImpl) GetCurrencyUsageStats(ctx context.Context) (*currency.CurrencyUsageStats, error) {
	// Get total currencies
	totalCurrencies, err := s.repository.CountCurrencies(ctx, &currency.CurrencyFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to count currencies: %w", err)
	}

	// Get active currencies
	activeCurrencies, err := s.repository.CountCurrencies(ctx, &currency.CurrencyFilter{
		IsActive: &[]bool{true}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count active currencies: %w", err)
	}

	// Get total transactions
	totalTransactions, err := s.repository.CountTransactions(ctx, &currency.TransactionFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get total volume
	transactions, err := s.repository.ListTransactions(ctx, &currency.TransactionFilter{
		Status: currency.TransactionStatusCompleted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for volume: %w", err)
	}

	totalVolume := 0.0
	currencyDistribution := make(map[string]int64)

	for _, tx := range transactions {
		totalVolume += tx.BaseAmount
		currencyDistribution[tx.Currency]++
	}

	// Find most used currency
	mostUsedCurrency := ""
	maxCount := int64(0)

	for currency, count := range currencyDistribution {
		if count > maxCount {
			maxCount = count
			mostUsedCurrency = currency
		}
	}

	// Get exchange rate count (using ListExchangeRates for now since CountExchangeRates doesn't exist)
	exchangeRates, err := s.repository.ListExchangeRates(ctx, &currency.ExchangeRateFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to count exchange rates: %w", err)
	}

	return &currency.CurrencyUsageStats{
		TotalCurrencies:      totalCurrencies,
		ActiveCurrencies:     activeCurrencies,
		TotalTransactions:    int64(totalTransactions),
		TotalVolume:          totalVolume,
		MostUsedCurrency:     mostUsedCurrency,
		CurrencyDistribution: currencyDistribution,
		ExchangeRateCount:    len(exchangeRates),
		LastUpdated:          time.Now(),
	}, nil
}

// GetMonthlyRevenueTrends calculates monthly revenue trends
func (s *AnalyticsServiceImpl) GetMonthlyRevenueTrends(ctx context.Context, months int) (map[string]float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	filter := &currency.TransactionFilter{
		Status:   currency.TransactionStatusCompleted,
		FromDate: &startDate,
		ToDate:   &endDate,
	}

	transactions, err := s.repository.ListTransactions(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get transactions for monthly trends")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	monthlyRevenue := make(map[string]float64)

	for _, tx := range transactions {
		monthKey := tx.CreatedAt.Format("2006-01")
		monthlyRevenue[monthKey] += tx.BaseAmount
	}

	return monthlyRevenue, nil
}

// GetTopCurrenciesByRevenue returns top currencies by revenue
func (s *AnalyticsServiceImpl) GetTopCurrenciesByRevenue(ctx context.Context, limit int) ([]*currency.CurrencyRevenue, error) {
	revenueByCurrency, err := s.GetRevenueByCurrency(ctx, &currency.TransactionFilter{
		Status: currency.TransactionStatusCompleted,
	})
	if err != nil {
		return nil, err
	}

	// Convert to slice and sort
	var currencyRevenues []*currency.CurrencyRevenue
	for currCode, revenue := range revenueByCurrency {
		currencyRevenues = append(currencyRevenues, &currency.CurrencyRevenue{
			Currency: currCode,
			Revenue:  revenue,
		})
	}

	// Simple sort (in production, use proper sorting)
	if len(currencyRevenues) > limit {
		currencyRevenues = currencyRevenues[:limit]
	}

	return currencyRevenues, nil
}

// GetTransactionTypeAnalytics returns analytics by transaction type
func (s *AnalyticsServiceImpl) GetTransactionTypeAnalytics(ctx context.Context, filter *currency.TransactionFilter) (map[string]*currency.TransactionTypeStats, error) {
	if filter == nil {
		filter = &currency.TransactionFilter{}
	}

	transactions, err := s.repository.ListTransactions(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get transactions for type analytics")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	typeStats := make(map[string]*currency.TransactionTypeStats)

	for _, tx := range transactions {
		typeKey := string(tx.Type)

		if _, exists := typeStats[typeKey]; !exists {
			typeStats[typeKey] = &currency.TransactionTypeStats{
				Type:     tx.Type,
				Count:    0,
				Amount:   0.0,
				Currency: tx.Currency,
			}
		}

		stats := typeStats[typeKey]
		stats.Count++
		stats.Amount += tx.BaseAmount
	}

	return typeStats, nil
}

package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RateProviderType defines external rate providers
type RateProviderType string

const (
	ProviderOpenExchange RateProviderType = "openexchangerates"
	ProviderFixer        RateProviderType = "fixer"
	ProviderXE           RateProviderType = "xe"
	ProviderInternal     RateProviderType = "internal"
)

// RealTimeExchangeService provides real-time exchange rates
type RealTimeExchangeService struct {
	provider     RateProviderType
	apiKey       string
	baseCurrency string
	rates        map[string]float64
	lastUpdate   time.Time
	cacheTTL     time.Duration
	mu           sync.RWMutex
	logger       *logrus.Logger
	httpClient   *http.Client
}

// ExchangeRateConfig configures the exchange rate service
type ExchangeRateConfig struct {
	Provider     RateProviderType
	APIKey       string
	BaseCurrency string
	CacheTTL     time.Duration
}

// NewRealTimeExchangeService creates a new exchange rate service
func NewRealTimeExchangeService(config ExchangeRateConfig, logger *logrus.Logger) *RealTimeExchangeService {
	if config.CacheTTL == 0 {
		config.CacheTTL = 15 * time.Minute
	}
	if config.BaseCurrency == "" {
		config.BaseCurrency = "USD"
	}

	svc := &RealTimeExchangeService{
		provider:     config.Provider,
		apiKey:       config.APIKey,
		baseCurrency: config.BaseCurrency,
		rates:        make(map[string]float64),
		cacheTTL:     config.CacheTTL,
		logger:       logger,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	// Initialize with default rates
	svc.initDefaultRates()

	return svc
}

func (s *RealTimeExchangeService) initDefaultRates() {
	s.rates = map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"GBP": 0.79,
		"JPY": 149.50,
		"CNY": 7.24,
		"INR": 83.12,
		"BRL": 4.97,
		"CAD": 1.36,
		"AUD": 1.53,
		"CHF": 0.88,
		"SGD": 1.34,
		"HKD": 7.82,
		"KRW": 1320.0,
		"MXN": 17.15,
		"ZAR": 18.50,
	}
	s.lastUpdate = time.Now()
}

// GetRate returns the exchange rate for a currency pair
func (s *RealTimeExchangeService) GetRate(ctx context.Context, from, to string) (float64, error) {
	s.mu.RLock()
	needsRefresh := time.Since(s.lastUpdate) > s.cacheTTL
	s.mu.RUnlock()

	if needsRefresh {
		if err := s.RefreshRates(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to refresh rates, using cached")
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	fromRate, fromOK := s.rates[from]
	toRate, toOK := s.rates[to]

	if !fromOK {
		return 0, fmt.Errorf("unknown currency: %s", from)
	}
	if !toOK {
		return 0, fmt.Errorf("unknown currency: %s", to)
	}

	// Convert through base currency
	return toRate / fromRate, nil
}

// Convert converts an amount between currencies
func (s *RealTimeExchangeService) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}

// RefreshRates fetches latest rates from provider
func (s *RealTimeExchangeService) RefreshRates(ctx context.Context) error {
	var err error

	switch s.provider {
	case ProviderOpenExchange:
		err = s.fetchOpenExchangeRates(ctx)
	case ProviderFixer:
		err = s.fetchFixerRates(ctx)
	case ProviderInternal:
		// Use internal rates, no fetch needed
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", s.provider)
	}

	if err != nil {
		return err
	}

	s.mu.Lock()
	s.lastUpdate = time.Now()
	s.mu.Unlock()

	s.logger.Info("Exchange rates refreshed")
	return nil
}

func (s *RealTimeExchangeService) fetchOpenExchangeRates(ctx context.Context) error {
	url := fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s&base=%s",
		s.apiKey, s.baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	s.mu.Lock()
	for currency, rate := range result.Rates {
		s.rates[currency] = rate
	}
	s.mu.Unlock()

	return nil
}

func (s *RealTimeExchangeService) fetchFixerRates(ctx context.Context) error {
	url := fmt.Sprintf("http://data.fixer.io/api/latest?access_key=%s&base=%s",
		s.apiKey, s.baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool               `json:"success"`
		Rates   map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("fixer API returned error")
	}

	s.mu.Lock()
	for currency, rate := range result.Rates {
		s.rates[currency] = rate
	}
	s.mu.Unlock()

	return nil
}

// GetAllRates returns all cached rates
func (s *RealTimeExchangeService) GetAllRates() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rates := make(map[string]float64)
	for k, v := range s.rates {
		rates[k] = v
	}
	return rates
}

// GetSupportedCurrencies returns list of supported currencies
func (s *RealTimeExchangeService) GetSupportedCurrencies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currencies := make([]string, 0, len(s.rates))
	for c := range s.rates {
		currencies = append(currencies, c)
	}
	return currencies
}

// LastUpdateTime returns when rates were last updated
func (s *RealTimeExchangeService) LastUpdateTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastUpdate
}

// ExchangeRateHistory stores historical exchange rates
type ExchangeRateHistory struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	FromCurr  string    `json:"from_currency" gorm:"index"`
	ToCurr    string    `json:"to_currency" gorm:"index"`
	Rate      float64   `json:"rate"`
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
}

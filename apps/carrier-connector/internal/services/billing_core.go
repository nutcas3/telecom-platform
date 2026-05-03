package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/sirupsen/logrus"
)

// BillingServiceImpl handles multi-currency billing operations
type BillingServiceImpl struct {
	repository      currency.Repository
	exchangeService currency.ExchangeRateService
	logger          *logrus.Logger
	baseCurrency    string
}

// NewBillingService creates a new billing service
func NewBillingService(repository currency.Repository, exchangeService currency.ExchangeRateService, logger *logrus.Logger, baseCurrency string) *BillingServiceImpl {
	return &BillingServiceImpl{
		repository:      repository,
		exchangeService: exchangeService,
		logger:          logger,
		baseCurrency:    baseCurrency,
	}
}

// ProcessBilling processes a billing request in multi-currency context
func (s *BillingServiceImpl) ProcessBilling(ctx context.Context, req *currency.BillingRequest) (*currency.BillingResponse, error) {
	// Validate request
	if err := s.validateBillingRequest(req); err != nil {
		return nil, fmt.Errorf("invalid billing request: %w", err)
	}

	// Convert to base currency if needed
	baseAmount := req.Amount
	exchangeRate := 1.0

	if req.Currency != s.baseCurrency {
		conversion, err := s.exchangeService.ConvertAmount(ctx, req.Amount, req.Currency, s.baseCurrency)
		if err != nil {
			s.logger.WithError(err).Error("Failed to convert currency for billing")
			return nil, fmt.Errorf("currency conversion failed: %w", err)
		}
		baseAmount = conversion.ConvertedAmount
		exchangeRate = conversion.ExchangeRate
	}

	// Create transaction
	transaction := &currency.Transaction{
		ID:             uuid.New().String(),
		ProfileID:      req.ProfileID,
		SubscriptionID: req.SubscriptionID,
		Type:           currency.TransactionTypeSubscription,
		Amount:         req.Amount,
		Currency:       req.Currency,
		BaseAmount:     baseAmount,
		BaseCurrency:   s.baseCurrency,
		ExchangeRate:   exchangeRate,
		Description:    req.Description,
		Status:         currency.TransactionStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save transaction
	if err := s.repository.CreateTransaction(ctx, transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create billing transaction")
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Process payment (in real implementation, this would integrate with payment gateway)
	transaction.Status = currency.TransactionStatusCompleted
	if err := s.repository.UpdateTransaction(ctx, transaction); err != nil {
		s.logger.WithError(err).Error("Failed to update transaction status")
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"transaction_id": transaction.ID,
		"profile_id":     req.ProfileID,
		"amount":         req.Amount,
		"currency":       req.Currency,
		"base_amount":    baseAmount,
		"base_currency":  s.baseCurrency,
	}).Info("Billing processed successfully")

	return &currency.BillingResponse{
		TransactionID: transaction.ID,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		BaseAmount:    transaction.BaseAmount,
		BaseCurrency:  transaction.BaseCurrency,
		ExchangeRate:  transaction.ExchangeRate,
		Status:        string(transaction.Status),
		ProcessedAt:   time.Now(),
	}, nil
}

// ConvertAmount converts an amount between currencies
func (s *BillingServiceImpl) ConvertAmount(ctx context.Context, req *currency.CurrencyConversionRequest) (*currency.CurrencyConversionResponse, error) {
	return s.exchangeService.ConvertAmount(ctx, req.Amount, req.FromCurrency, req.ToCurrency)
}

// GetBillingHistory retrieves billing history for a profile
func (s *BillingServiceImpl) GetBillingHistory(ctx context.Context, profileID string, filter *currency.TransactionFilter) ([]*currency.Transaction, error) {
	if filter == nil {
		filter = &currency.TransactionFilter{}
	}

	filter.ProfileID = profileID

	transactions, err := s.repository.ListTransactions(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get billing history")
		return nil, fmt.Errorf("failed to get billing history: %w", err)
	}

	return transactions, nil
}

// validateBillingRequest validates a billing request
func (s *BillingServiceImpl) validateBillingRequest(req *currency.BillingRequest) error {
	if req.ProfileID == "" {
		return fmt.Errorf("profile ID is required")
	}
	if req.SubscriptionID == "" {
		return fmt.Errorf("subscription ID is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.BillingDate.IsZero() {
		req.BillingDate = time.Now()
	}

	// Validate currency
	if err := s.exchangeService.ValidateCurrencyPair(context.Background(), req.Currency, s.baseCurrency); err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	return nil
}

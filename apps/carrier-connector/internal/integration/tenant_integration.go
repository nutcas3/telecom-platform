package integration

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/services"
	"github.com/sirupsen/logrus"
)

// TenantAwareCurrencyRepository wraps the currency repository with tenant isolation
type TenantAwareCurrencyRepository struct {
	baseRepo currency.Repository
	tenantID string
	logger   *logrus.Logger
}

// NewTenantAwareCurrencyRepository creates a new tenant-aware currency repository
func NewTenantAwareCurrencyRepository(baseRepo currency.Repository, tenantID string, logger *logrus.Logger) currency.Repository {
	return &TenantAwareCurrencyRepository{
		baseRepo: baseRepo,
		tenantID: tenantID,
		logger:   logger,
	}
}

// WithTenant creates a new instance with the specified tenant ID
func (r *TenantAwareCurrencyRepository) WithTenant(tenantID string) currency.Repository {
	return &TenantAwareCurrencyRepository{
		baseRepo: r.baseRepo,
		tenantID: tenantID,
		logger:   r.logger,
	}
}

// Tenant-aware currency repository methods
func (r *TenantAwareCurrencyRepository) CreateCurrency(ctx context.Context, currency *currency.Currency) error {
	// Add tenant context
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.CreateCurrency(ctx, currency)
}

func (r *TenantAwareCurrencyRepository) GetCurrency(ctx context.Context, code string) (*currency.Currency, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.GetCurrency(ctx, code)
}

func (r *TenantAwareCurrencyRepository) UpdateCurrency(ctx context.Context, currency *currency.Currency) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.UpdateCurrency(ctx, currency)
}

func (r *TenantAwareCurrencyRepository) DeleteCurrency(ctx context.Context, code string) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.DeleteCurrency(ctx, code)
}

func (r *TenantAwareCurrencyRepository) ListCurrencies(ctx context.Context, filter *currency.CurrencyFilter) ([]*currency.Currency, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.ListCurrencies(ctx, filter)
}

func (r *TenantAwareCurrencyRepository) CountCurrencies(ctx context.Context, filter *currency.CurrencyFilter) (int, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.CountCurrencies(ctx, filter)
}

func (r *TenantAwareCurrencyRepository) CreateExchangeRate(ctx context.Context, rate *currency.ExchangeRate) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.CreateExchangeRate(ctx, rate)
}

func (r *TenantAwareCurrencyRepository) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*currency.ExchangeRate, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.GetExchangeRate(ctx, fromCurrency, toCurrency)
}

func (r *TenantAwareCurrencyRepository) UpdateExchangeRate(ctx context.Context, rate *currency.ExchangeRate) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.UpdateExchangeRate(ctx, rate)
}

func (r *TenantAwareCurrencyRepository) DeleteExchangeRate(ctx context.Context, id string) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.DeleteExchangeRate(ctx, id)
}

func (r *TenantAwareCurrencyRepository) ListExchangeRates(ctx context.Context, filter *currency.ExchangeRateFilter) ([]*currency.ExchangeRate, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.ListExchangeRates(ctx, filter)
}

func (r *TenantAwareCurrencyRepository) GetLatestExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*currency.ExchangeRate, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.GetLatestExchangeRate(ctx, fromCurrency, toCurrency)
}

func (r *TenantAwareCurrencyRepository) CreateTransaction(ctx context.Context, transaction *currency.Transaction) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.CreateTransaction(ctx, transaction)
}

func (r *TenantAwareCurrencyRepository) GetTransaction(ctx context.Context, id string) (*currency.Transaction, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.GetTransaction(ctx, id)
}

func (r *TenantAwareCurrencyRepository) UpdateTransaction(ctx context.Context, transaction *currency.Transaction) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.UpdateTransaction(ctx, transaction)
}

func (r *TenantAwareCurrencyRepository) DeleteTransaction(ctx context.Context, id string) error {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.DeleteTransaction(ctx, id)
}

func (r *TenantAwareCurrencyRepository) ListTransactions(ctx context.Context, filter *currency.TransactionFilter) ([]*currency.Transaction, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.ListTransactions(ctx, filter)
}

func (r *TenantAwareCurrencyRepository) CountTransactions(ctx context.Context, filter *currency.TransactionFilter) (int, error) {
	ctx = repository.SetTenantInContext(ctx, r.tenantID)
	return r.baseRepo.CountTransactions(ctx, filter)
}

// TenantIntegrationManager manages tenant integration across all services
type TenantIntegrationManager struct {
	tenantService   *services.TenantServiceImpl
	currencyService currency.BillingService
	logger          *logrus.Logger
}

// NewTenantIntegrationManager creates a new tenant integration manager
func NewTenantIntegrationManager(
	tenantService *services.TenantServiceImpl,
	currencyService currency.BillingService,
	logger *logrus.Logger,
) *TenantIntegrationManager {
	return &TenantIntegrationManager{
		tenantService:   tenantService,
		currencyService: currencyService,
		logger:          logger,
	}
}

// GetTenantAwareServices returns tenant-aware service instances
func (m *TenantIntegrationManager) GetTenantAwareServices(ctx context.Context, tenantID string) (*TenantAwareServices, error) {
	// Validate tenant
	tenantCtx, err := m.tenantService.ValidateTenantAccess(ctx, tenantID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate tenant: %w", err)
	}

	// Get tenant configuration
	config, err := m.tenantService.GetTenantConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Create tenant-aware services
	services := &TenantAwareServices{
		TenantID:        tenantID,
		TenantContext:   tenantCtx,
		Config:          config,
		CurrencyService: m.wrapCurrencyService(tenantID),
	}

	return services, nil
}

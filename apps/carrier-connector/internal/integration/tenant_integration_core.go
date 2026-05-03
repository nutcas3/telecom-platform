package integration

import (
	"context"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/services"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"github.com/sirupsen/logrus"
)

// TenantAwareCurrencyService wraps a currency service with tenant isolation
type TenantAwareCurrencyService struct {
	tenantID       string
	billingService currency.BillingService
	logger         *logrus.Logger
}

// ProcessBilling processes billing requests with tenant isolation
func (tacs *TenantAwareCurrencyService) ProcessBilling(ctx context.Context, req *currency.BillingRequest) (*currency.BillingResponse, error) {
	// Log the billing operation with tenant context
	tacs.logger.WithFields(logrus.Fields{
		"tenant_id":  tacs.tenantID,
		"profile_id": req.ProfileID,
		"amount":     req.Amount,
		"currency":   req.Currency,
	}).Info("Processing tenant billing request")

	// Inject tenant ID into context for downstream isolation
	ctx = context.WithValue(ctx, "tenant_id", tacs.tenantID)
	return tacs.billingService.ProcessBilling(ctx, req)
}

// ConvertAmount converts currency amounts with tenant isolation
func (tacs *TenantAwareCurrencyService) ConvertAmount(ctx context.Context, req *currency.CurrencyConversionRequest) (*currency.CurrencyConversionResponse, error) {
	// Log the conversion operation with tenant context
	tacs.logger.WithFields(logrus.Fields{
		"tenant_id":     tacs.tenantID,
		"from_currency": req.FromCurrency,
		"to_currency":   req.ToCurrency,
		"amount":        req.Amount,
	}).Info("Processing tenant currency conversion")

	// Inject tenant ID into context for downstream isolation
	ctx = context.WithValue(ctx, "tenant_id", tacs.tenantID)
	return tacs.billingService.ConvertAmount(ctx, req)
}

// GetBillingHistory retrieves billing history with tenant isolation
func (tacs *TenantAwareCurrencyService) GetBillingHistory(ctx context.Context, profileID string, filter *currency.TransactionFilter) ([]*currency.Transaction, error) {
	// Log the history retrieval with tenant context
	tacs.logger.WithFields(logrus.Fields{
		"tenant_id":  tacs.tenantID,
		"profile_id": profileID,
	}).Info("Retrieving tenant billing history")
	// Inject tenant ID into context so the repository layer filters by tenant
	ctx = context.WithValue(ctx, "tenant_id", tacs.tenantID)
	return tacs.billingService.GetBillingHistory(ctx, profileID, filter)
}

// CalculateTotalBilling calculates total billing with tenant isolation
func (tacs *TenantAwareCurrencyService) CalculateTotalBilling(ctx context.Context, profileID string, fromDate, toDate time.Time) (*currency.BillingSummary, error) {
	// Log the calculation
	tacs.logger.WithFields(logrus.Fields{
		"tenant_id":  tacs.tenantID,
		"profile_id": profileID,
		"from_date":  fromDate,
		"to_date":    toDate,
	}).Info("Calculating tenant total billing")

	// Inject tenant ID into context so the repository layer scopes calculations
	ctx = context.WithValue(ctx, "tenant_id", tacs.tenantID)
	return tacs.billingService.CalculateTotalBilling(ctx, profileID, fromDate, toDate)
}

// TenantAwareServices provides tenant-aware service instances
type TenantAwareServices struct {
	TenantID        string
	TenantContext   *tenant.TenantContext
	Config          *tenant.TenantConfig
	CurrencyService currency.BillingService
}

// wrapCurrencyService creates a tenant-aware currency service
func (m *TenantIntegrationManager) wrapCurrencyService(tenantID string) currency.BillingService {
	return &TenantAwareCurrencyService{
		tenantID:       tenantID,
		billingService: m.currencyService,
		logger:         m.logger,
	}
}

// TenantResourceQuotaChecker checks resource quotas before operations
type TenantResourceQuotaChecker struct {
	tenantService *services.TenantServiceImpl
	logger        *logrus.Logger
}

// NewTenantResourceQuotaChecker creates a new quota checker
func NewTenantResourceQuotaChecker(tenantService *services.TenantServiceImpl, logger *logrus.Logger) *TenantResourceQuotaChecker {
	return &TenantResourceQuotaChecker{
		tenantService: tenantService,
		logger:        logger,
	}
}

// CheckQuota checks if tenant has sufficient quota for a resource operation
func (c *TenantResourceQuotaChecker) CheckQuota(ctx context.Context, tenantID, resourceType string, count int) error {
	return c.tenantService.CheckQuota(ctx, tenantID, resourceType, count)
}

// UpdateUsage updates resource usage after an operation
func (c *TenantResourceQuotaChecker) UpdateUsage(ctx context.Context, tenantID, resourceType string, count int) error {
	return c.tenantService.UpdateUsage(ctx, tenantID, resourceType, count)
}

// TenantEventLogger logs tenant events for audit purposes
type TenantEventLogger struct {
	tenantService *services.TenantServiceImpl
	logger        *logrus.Logger
}

// NewTenantEventLogger creates a new tenant event logger
func NewTenantEventLogger(tenantService *services.TenantServiceImpl, logger *logrus.Logger) *TenantEventLogger {
	return &TenantEventLogger{
		tenantService: tenantService,
		logger:        logger,
	}
}

// LogResourceAccess logs resource access events
func (l *TenantEventLogger) LogResourceAccess(ctx context.Context, tenantID, userID, resourceType, resourceID, action string) {
	event := &tenant.TenantEvent{
		ID:        id.GeneratePrefixed("tnt"),
		TenantID:  tenantID,
		UserID:    userID,
		EventType: tenant.TenantEventType("resource_access"),
		EventData: map[string]any{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"action":        action,
		},
		Timestamp: time.Now(),
	}

	if err := l.tenantService.LogTenantEvent(ctx, event); err != nil {
		l.logger.WithError(err).Error("Failed to log resource access event")
	}
}

// LogQuotaViolation logs quota violation events
func (l *TenantEventLogger) LogQuotaViolation(ctx context.Context, tenantID, resourceType string, usage, limit int) {
	event := &tenant.TenantEvent{
		ID:        id.GeneratePrefixed("tnt"),
		TenantID:  tenantID,
		UserID:    "",
		EventType: tenant.TenantEventQuotaExceeded,
		EventData: map[string]any{
			"resource_type": resourceType,
			"usage":         usage,
			"limit":         limit,
		},
		Timestamp: time.Now(),
	}

	if err := l.tenantService.LogTenantEvent(ctx, event); err != nil {
		l.logger.WithError(err).Error("Failed to log quota violation event")
	}
}

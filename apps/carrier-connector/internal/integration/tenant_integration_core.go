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

// TenantAwareServices provides tenant-aware service instances
type TenantAwareServices struct {
	TenantID        string
	TenantContext   *tenant.TenantContext
	Config          *tenant.TenantConfig
	CurrencyService currency.BillingService
}

// wrapCurrencyService creates a tenant-aware currency service
func (m *TenantIntegrationManager) wrapCurrencyService(tenantID string) currency.BillingService {
	// TODO: Implement tenant isolation for currency service
	// The tenantID parameter should be used to filter currency operations by tenant
	// For now, return the base service - implementation needed for multi-tenant isolation
	_ = tenantID // Suppress unused parameter warning until implementation is complete
	return m.currencyService
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

package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"github.com/sirupsen/logrus"
)

// TenantServiceImpl implements the tenant service interface
type TenantServiceImpl struct {
	repository  tenant.Repository
	rateLimiter tenant.RateLimiter
	logger      *logrus.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(
	repository tenant.Repository,
	rateLimiter tenant.RateLimiter,
	logger *logrus.Logger,
) tenant.Service {
	return &TenantServiceImpl{
		repository:  repository,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// CreateTenant creates a new tenant
func (s *TenantServiceImpl) CreateTenant(ctx context.Context, req *tenant.CreateTenantRequest) (*tenant.Tenant, error) {
	// Validate request
	if err := s.validateCreateTenantRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if domain already exists
	existing, err := s.repository.GetTenantByDomain(ctx, req.Domain)
	if err == nil && existing != nil {
		return nil, errors.New("domain already exists")
	}

	// Create tenant
	tenant := &tenant.Tenant{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Domain:      req.Domain,
		Status:      tenant.TenantStatusActive,
		Plan:        req.Plan,
		MaxUsers:    req.MaxUsers,
		MaxProfiles: req.MaxProfiles,
		MaxCarriers: req.MaxCarriers,
		Settings:    req.Settings,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set default settings if not provided
	if tenant.Settings == nil {
		tenant.Settings = s.getDefaultSettings(req.Plan)
	}

	// Save tenant
	if err := s.repository.CreateTenant(ctx, tenant); err != nil {
		s.logger.WithError(err).Error("Failed to create tenant")
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create initial configuration
	config := &tenant.TenantConfig{
		TenantID: tenant.ID,
		Config:   make(map[string]interface{}),
		Settings: tenant.Settings,
		Quotas:   s.getDefaultQuotas(req.Plan),
		Features: s.getDefaultFeatures(req.Plan),
	}

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		s.logger.WithError(err).Error("Failed to create tenant config")
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenant.ID,
		"name":      tenant.Name,
		"domain":    tenant.Domain,
	}).Info("Tenant created successfully")

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *TenantServiceImpl) GetTenant(ctx context.Context, id string) (*tenant.Tenant, error) {
	tenant, err := s.repository.GetTenant(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("tenant_id", id).Error("Failed to get tenant")
		return nil, err
	}

	return tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (s *TenantServiceImpl) GetTenantByDomain(ctx context.Context, domain string) (*tenant.Tenant, error) {
	tenant, err := s.repository.GetTenantByDomain(ctx, domain)
	if err != nil {
		s.logger.WithError(err).WithField("domain", domain).Error("Failed to get tenant by domain")
		return nil, err
	}

	return tenant, nil
}

// UpdateTenant updates an existing tenant
func (s *TenantServiceImpl) UpdateTenant(ctx context.Context, id string, req *tenant.UpdateTenantRequest) (*tenant.Tenant, error) {
	// Get existing tenant
	tenant, err := s.repository.GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Plan != nil {
		tenant.Plan = *req.Plan
	}
	if req.MaxUsers != nil {
		tenant.MaxUsers = *req.MaxUsers
	}
	if req.MaxProfiles != nil {
		tenant.MaxProfiles = *req.MaxProfiles
	}
	if req.MaxCarriers != nil {
		tenant.MaxCarriers = *req.MaxCarriers
	}
	if req.Settings != nil {
		tenant.Settings = req.Settings
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	tenant.UpdatedAt = time.Now()

	// Save tenant
	if err := s.repository.UpdateTenant(ctx, tenant); err != nil {
		s.logger.WithError(err).Error("Failed to update tenant")
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	s.logger.WithField("tenant_id", tenant.ID).Info("Tenant updated successfully")

	return tenant, nil
}

// DeleteTenant deletes a tenant
func (s *TenantServiceImpl) DeleteTenant(ctx context.Context, id string) error {
	// Delete tenant
	if err := s.repository.DeleteTenant(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete tenant")
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.logger.WithField("tenant_id", id).Info("Tenant deleted successfully")

	return nil
}

// ListTenants lists tenants with filtering
func (s *TenantServiceImpl) ListTenants(ctx context.Context, filter *tenant.TenantFilter) ([]*tenant.Tenant, error) {
	tenants, err := s.repository.ListTenants(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list tenants")
		return nil, err
	}

	return tenants, nil
}

// Helper methods
func (s *TenantServiceImpl) validateCreateTenantRequest(req *tenant.CreateTenantRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Domain == "" {
		return errors.New("domain is required")
	}
	if req.Plan == "" {
		return errors.New("plan is required")
	}
	return nil
}

func (s *TenantServiceImpl) getDefaultSettings(plan tenant.TenantPlan) *tenant.TenantSettings {
	settings := &tenant.TenantSettings{
		DefaultCurrency:       "USD",
		SupportedCurrencies:   []string{"USD", "EUR", "GBP"},
		APIRateLimitPerMinute: 60,
		APIRateLimitPerHour:   1000,
		SessionTimeout:        120, // 2 hours
		DataRetentionDays:     90,
		ComplianceRegions:     []string{"US", "EU"},
	}

	switch plan {
	case tenant.TenantPlanFree:
		settings.EnableMultiCurrency = false
		settings.EnableAdvancedAnalytics = false
		settings.EnableAPIAccess = true
		settings.EnableWebhooks = false
		settings.Require2FA = false
	case tenant.TenantPlanBasic:
		settings.EnableMultiCurrency = true
		settings.EnableAdvancedAnalytics = false
		settings.EnableAPIAccess = true
		settings.EnableWebhooks = false
		settings.Require2FA = false
	case tenant.TenantPlanPro:
		settings.EnableMultiCurrency = true
		settings.EnableAdvancedAnalytics = true
		settings.EnableAPIAccess = true
		settings.EnableWebhooks = true
		settings.Require2FA = true
	case tenant.TenantPlanEnterprise:
		settings.EnableMultiCurrency = true
		settings.EnableAdvancedAnalytics = true
		settings.EnableAPIAccess = true
		settings.EnableWebhooks = true
		settings.Require2FA = true
		settings.APIRateLimitPerMinute = 1000
		settings.APIRateLimitPerHour = 10000
	}

	return settings
}

func (s *TenantServiceImpl) getDefaultQuotas(plan tenant.TenantPlan) []tenant.ResourceQuota {
	quotas := []tenant.ResourceQuota{}

	switch plan {
	case tenant.TenantPlanFree:
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "users", Limit: 5, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "profiles", Limit: 100, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "carriers", Limit: 3, Period: "monthly"})
	case tenant.TenantPlanBasic:
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "users", Limit: 25, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "profiles", Limit: 1000, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "carriers", Limit: 10, Period: "monthly"})
	case tenant.TenantPlanPro:
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "users", Limit: 100, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "profiles", Limit: 10000, Period: "monthly"})
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "carriers", Limit: 50, Period: "monthly"})
	case tenant.TenantPlanEnterprise:
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "users", Limit: -1, Period: "monthly"})    // Unlimited
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "profiles", Limit: -1, Period: "monthly"}) // Unlimited
		quotas = append(quotas, tenant.ResourceQuota{ResourceType: "carriers", Limit: -1, Period: "monthly"}) // Unlimited
	}

	return quotas
}

func (s *TenantServiceImpl) getDefaultFeatures(plan tenant.TenantPlan) map[string]bool {
	features := map[string]bool{
		"multi_currency":     false,
		"advanced_analytics": false,
		"api_access":         true,
		"webhooks":           false,
		"custom_branding":    false,
		"priority_support":   false,
	}

	switch plan {
	case tenant.TenantPlanBasic:
		features["multi_currency"] = true
	case tenant.TenantPlanPro:
		features["multi_currency"] = true
		features["advanced_analytics"] = true
		features["webhooks"] = true
		features["custom_branding"] = true
	case tenant.TenantPlanEnterprise:
		features["multi_currency"] = true
		features["advanced_analytics"] = true
		features["webhooks"] = true
		features["custom_branding"] = true
		features["priority_support"] = true
	}

	return features
}

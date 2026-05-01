package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// GetTenantConfig retrieves tenant configuration
func (s *TenantServiceImpl) GetTenantConfig(ctx context.Context, tenantID string) (*tenant.TenantConfig, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config, nil
}

// UpdateTenantConfig updates tenant configuration
func (s *TenantServiceImpl) UpdateTenantConfig(ctx context.Context, tenantID string, config *tenant.TenantConfig) error {
	config.TenantID = tenantID

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update tenant config: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant config updated successfully")

	return nil
}

// GetTenantSettings retrieves tenant settings
func (s *TenantServiceImpl) GetTenantSettings(ctx context.Context, tenantID string) (*tenant.TenantSettings, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config.Settings, nil
}

// UpdateTenantSettings updates tenant settings
func (s *TenantServiceImpl) UpdateTenantSettings(ctx context.Context, tenantID string, settings *tenant.TenantSettings) error {
	// Get current config
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Update settings
	config.Settings = settings

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant settings updated successfully")

	return nil
}

// GetTenantFeatures retrieves tenant features
func (s *TenantServiceImpl) GetTenantFeatures(ctx context.Context, tenantID string) (map[string]bool, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config.Features, nil
}

// UpdateTenantFeatures updates tenant features
func (s *TenantServiceImpl) UpdateTenantFeatures(ctx context.Context, tenantID string, features map[string]bool) error {
	// Get current config
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Update features
	config.Features = features

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update tenant features: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant features updated successfully")

	return nil
}

// GetTenantQuotas retrieves tenant quotas
func (s *TenantServiceImpl) GetTenantQuotas(ctx context.Context, tenantID string) ([]tenant.ResourceQuota, error) {
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}

	return config.Quotas, nil
}

// UpdateTenantQuotas updates tenant quotas
func (s *TenantServiceImpl) UpdateTenantQuotas(ctx context.Context, tenantID string, quotas []tenant.ResourceQuota) error {
	// Get current config
	config, err := s.repository.GetConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	// Update quotas
	config.Quotas = quotas

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update tenant quotas: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant quotas updated successfully")

	return nil
}

// ValidateTenantConfig validates tenant configuration
func (s *TenantServiceImpl) ValidateTenantConfig(ctx context.Context, config *tenant.TenantConfig) error {
	if config.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	if config.Settings == nil {
		return fmt.Errorf("settings are required")
	}

	if config.Quotas == nil {
		return fmt.Errorf("quotas are required")
	}

	if config.Features == nil {
		return fmt.Errorf("features are required")
	}

	// Validate quotas
	for _, quota := range config.Quotas {
		if quota.ResourceType == "" {
			return fmt.Errorf("quota resource type is required")
		}
		if quota.Limit < -1 {
			return fmt.Errorf("quota limit must be -1 (unlimited) or greater")
		}
		if quota.Period == "" {
			return fmt.Errorf("quota period is required")
		}
	}

	return nil
}

// ResetTenantConfig resets tenant configuration to defaults
func (s *TenantServiceImpl) ResetTenantConfig(ctx context.Context, tenantID string) error {
	// Get tenant to determine plan
	tenant, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Create default config based on plan
	config := s.convertTenantConfig(tenantID, tenant.Settings, tenant.Plan)

	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to reset tenant config: %w", err)
	}

	s.logger.WithField("tenant_id", tenantID).Info("Tenant config reset to defaults successfully")

	return nil
}

// Helper functions
func (s *TenantServiceImpl) convertTenantConfig(tenantID string, settings *tenant.TenantSettings, plan tenant.TenantPlan) *tenant.TenantConfig {
	return &tenant.TenantConfig{
		TenantID: tenantID,
		Config:   make(map[string]interface{}),
		Settings: settings,
		Quotas:   s.getDefaultQuotas(plan),
		Features: s.getDefaultFeatures(plan),
	}
}

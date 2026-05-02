package services

import (
	"context"
	"errors"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
)

// HasPermission checks if user has permission
func (s *TenantServiceImpl) HasPermission(ctx context.Context, tenantID, userID, permission string) (bool, error) {
	// Get tenant user
	user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	// Simple permission check based on role
	switch user.Role {
	case tenant.TenantRoleOwner, tenant.TenantRoleAdmin:
		return true, nil
	case tenant.TenantRoleManager:
		return permission != "delete", nil
	case tenant.TenantRoleUser:
		return permission == "read", nil
	default:
		return false, nil
	}
}

func (s *TenantServiceImpl) validateCreateTenantUserRequest(req *tenant.CreateTenantUserRequest) error {
	if req.TenantID == "" {
		return errors.New("tenant ID is required")
	}
	if req.UserID == "" {
		return errors.New("user ID is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Role == "" {
		return errors.New("role is required")
	}
	return nil
}

func (s *TenantServiceImpl) convertTenantSettings(settings *tenant.TenantSettings) *tenant.TenantSettings {
	if settings == nil {
		return &tenant.TenantSettings{}
	}

	return &tenant.TenantSettings{
		DefaultCurrency:         settings.DefaultCurrency,
		SupportedCurrencies:     settings.SupportedCurrencies,
		EnableMultiCurrency:     settings.EnableMultiCurrency,
		EnableAdvancedAnalytics: settings.EnableAdvancedAnalytics,
		EnableAPIAccess:         settings.EnableAPIAccess,
		EnableWebhooks:          settings.EnableWebhooks,
		Require2FA:              settings.Require2FA,
		SessionTimeout:          settings.SessionTimeout,
		DataRetentionDays:       settings.DataRetentionDays,
		ComplianceRegions:       settings.ComplianceRegions,
	}
}

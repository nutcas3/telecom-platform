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

// AddUserToTenant adds a user to a tenant
func (s *TenantServiceImpl) AddUserToTenant(ctx context.Context, req *tenant.CreateTenantUserRequest) (*tenant.TenantUser, error) {
	// Validate request
	if err := s.validateCreateTenantUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existing, err := s.repository.GetTenantUser(ctx, req.TenantID, req.UserID)
	if err == nil && existing != nil {
		return nil, errors.New("user already exists in tenant")
	}

	// Create tenant user
	tenantUser := &tenant.TenantUser{
		ID:        uuid.New().String(),
		TenantID:  req.TenantID,
		UserID:    req.UserID,
		Email:     req.Email,
		Role:      req.Role,
		Status:    tenant.TenantUserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save tenant user
	if err := s.repository.CreateTenantUser(ctx, tenantUser); err != nil {
		s.logger.WithError(err).Error("Failed to create tenant user")
		return nil, fmt.Errorf("failed to add user to tenant: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": req.TenantID,
		"user_id":   req.UserID,
		"role":      req.Role,
	}).Info("User added to tenant successfully")

	return tenantUser, nil
}

// GetTenantUser retrieves a tenant user
func (s *TenantServiceImpl) GetTenantUser(ctx context.Context, tenantID, userID string) (*tenant.TenantUser, error) {
	user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"user_id":   userID,
		}).Error("Failed to get tenant user")
		return nil, err
	}

	return user, nil
}

// UpdateTenantUser updates a tenant user
func (s *TenantServiceImpl) UpdateTenantUser(ctx context.Context, tenantID, userID string, req *tenant.UpdateTenantUserRequest) (*tenant.TenantUser, error) {
	// Get existing user
	user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	user.UpdatedAt = time.Now()

	// Save user
	if err := s.repository.UpdateTenantUser(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to update tenant user")
		return nil, fmt.Errorf("failed to update tenant user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"user_id":   userID,
	}).Info("Tenant user updated successfully")

	return user, nil
}

// RemoveUserFromTenant removes a user from a tenant
func (s *TenantServiceImpl) RemoveUserFromTenant(ctx context.Context, tenantID, userID string) error {
	// Delete tenant user
	if err := s.repository.DeleteTenantUser(ctx, tenantID, userID); err != nil {
		s.logger.WithError(err).Error("Failed to remove user from tenant")
		return fmt.Errorf("failed to remove user from tenant: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"user_id":   userID,
	}).Info("User removed from tenant successfully")

	return nil
}

// ListTenantUsers lists tenant users with filtering
func (s *TenantServiceImpl) ListTenantUsers(ctx context.Context, filter *tenant.TenantUserFilter) ([]*tenant.TenantUser, error) {
	users, err := s.repository.ListTenantUsers(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list tenant users")
		return nil, err
	}

	return users, nil
}

// ValidateTenantAccess validates tenant access
func (s *TenantServiceImpl) ValidateTenantAccess(ctx context.Context, tenantID, userID string) (*tenant.TenantContext, error) {
	// Get tenant
	tenant, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check tenant status
	if tenant.Status != "active" {
		return nil, fmt.Errorf("tenant is not active")
	}

	// Get user
	user, err := s.repository.GetTenantUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found in tenant: %w", err)
	}

	// Check user status
	if user.Status != "active" {
		return nil, fmt.Errorf("user is not active")
	}

	// Create context
	context := &tenant.TenantContext{
		TenantID:    tenantID,
		UserID:      userID,
		Tenant:      tenant,
		User:        user,
		Roles:       []string{string(user.Role)},
		Permissions: []string{},
		Settings:    tenant.Settings,
		Quotas:      []tenant.ResourceQuota{},
		Features:    map[string]bool{},
	}

	return context, nil
}

// GetTenantContext retrieves tenant context
func (s *TenantServiceImpl) GetTenantContext(ctx context.Context, tenantID string) (*tenant.TenantContext, error) {
	// Get tenant
	tenant, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Create context with basic tenant info
	context := &tenant.TenantContext{
		TenantID:    tenantID,
		UserID:      "",
		Tenant:      tenant,
		User:        nil,
		Roles:       []string{},
		Permissions: []string{},
		Settings:    tenant.Settings,
		Quotas:      []tenant.ResourceQuota{},
		Features:    map[string]bool{},
	}

	return context, nil
}

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

// Helper methods
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

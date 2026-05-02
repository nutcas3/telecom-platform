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
	tenantRecord, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check tenant status
	if tenantRecord.Status != "active" {
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
	tenantCtx := &tenant.TenantContext{
		TenantID:   tenantID,
		TenantName: tenantRecord.Name,
		Plan:       tenantRecord.Plan,
		UserID:     userID,
		UserRole:   user.Role,
		Settings:   tenantRecord.Settings,
		Metadata:   tenantRecord.Metadata,
	}

	return tenantCtx, nil
}

// GetTenantContext retrieves tenant context
func (s *TenantServiceImpl) GetTenantContext(ctx context.Context, tenantID string) (*tenant.TenantContext, error) {
	// Get tenant
	tenantRecord, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Create context with basic tenant info
	tenantCtx := &tenant.TenantContext{
		TenantID:   tenantID,
		TenantName: tenantRecord.Name,
		Plan:       tenantRecord.Plan,
		UserID:     "",
		Settings:   tenantRecord.Settings,
		Metadata:   tenantRecord.Metadata,
	}

	return tenantCtx, nil
}

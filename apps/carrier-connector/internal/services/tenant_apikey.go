package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/tenant"
	"github.com/sirupsen/logrus"
)

// CreateAPIKey creates a new API key for a tenant
func (s *TenantServiceImpl) CreateAPIKey(ctx context.Context, tenantID string, req *tenant.CreateAPIKeyRequest) (*tenant.TenantAPIKey, string, error) {
	// Validate request
	if err := s.validateCreateAPIKeyRequest(req); err != nil {
		return nil, "", fmt.Errorf("validation failed: %w", err)
	}

	// Generate API key
	apiKey := s.generateAPIKey()
	keyHash, err := s.hashAPIKey(apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Create API key record
	apiKeyRecord := &tenant.TenantAPIKey{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		KeyHash:     keyHash,
		KeyPrefix:   apiKey[:8],
		Permissions: req.Permissions,
		RateLimit:   req.RateLimit,
		ExpiresAt:   req.ExpiresAt,
		Status:      tenant.APIKeyStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save API key
	if err := s.repository.CreateAPIKey(ctx, apiKeyRecord); err != nil {
		s.logger.WithError(err).Error("Failed to create API key")
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"key_id":    apiKeyRecord.ID,
		"name":      req.Name,
	}).Info("API key created successfully")

	return apiKeyRecord, apiKey, nil
}

// GetAPIKey retrieves an API key by ID
func (s *TenantServiceImpl) GetAPIKey(ctx context.Context, id string) (*tenant.TenantAPIKey, error) {
	apiKey, err := s.repository.GetAPIKey(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("key_id", id).Error("Failed to get API key")
		return nil, err
	}

	return apiKey, nil
}

// UpdateAPIKey updates an API key
func (s *TenantServiceImpl) UpdateAPIKey(ctx context.Context, id string, req *tenant.UpdateAPIKeyRequest) (*tenant.TenantAPIKey, error) {
	// Get existing API key
	apiKey, err := s.repository.GetAPIKey(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		apiKey.Name = *req.Name
	}
	if req.Permissions != nil {
		apiKey.Permissions = req.Permissions
	}
	if req.RateLimit != nil {
		apiKey.RateLimit = *req.RateLimit
	}
	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = req.ExpiresAt
	}
	if req.Status != nil {
		apiKey.Status = *req.Status
	}

	apiKey.UpdatedAt = time.Now()

	// Save API key
	if err := s.repository.UpdateAPIKey(ctx, apiKey); err != nil {
		s.logger.WithError(err).Error("Failed to update API key")
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	s.logger.WithField("key_id", id).Info("API key updated successfully")

	return apiKey, nil
}

// DeleteAPIKey deletes an API key
func (s *TenantServiceImpl) DeleteAPIKey(ctx context.Context, id string) error {
	// Delete API key
	if err := s.repository.DeleteAPIKey(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete API key")
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	s.logger.WithField("key_id", id).Info("API key deleted successfully")

	return nil
}

// ListAPIKeys lists API keys for a tenant
func (s *TenantServiceImpl) ListAPIKeys(ctx context.Context, tenantID string) ([]*tenant.TenantAPIKey, error) {
	apiKeys, err := s.repository.ListAPIKeys(ctx, tenantID)
	if err != nil {
		s.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to list API keys")
		return nil, err
	}

	return apiKeys, nil
}

// ValidateAPIKey validates an API key and returns the key record
func (s *TenantServiceImpl) ValidateAPIKey(ctx context.Context, key string) (*tenant.TenantAPIKey, error) {
	// Hash the provided key
	keyHash, err := s.hashAPIKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	// Look up API key by hash
	apiKey, err := s.repository.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	// Check if key is active
	if apiKey.Status != tenant.APIKeyStatusActive {
		return nil, errors.New("API key is not active")
	}

	// Check if key has expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New("API key has expired")
	}

	// Update last used timestamp
	now := time.Now()
	apiKey.LastUsed = &now
	if err := s.repository.UpdateAPIKey(ctx, apiKey); err != nil {
		s.logger.WithError(err).Error("Failed to update API key last used")
	}

	return apiKey, nil
}

// Helper methods
func (s *TenantServiceImpl) validateCreateAPIKeyRequest(req *tenant.CreateAPIKeyRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func (s *TenantServiceImpl) generateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure method if crypto/rand fails
		return uuid.New().String()
	}
	return "tk_" + hex.EncodeToString(bytes)
}

func (s *TenantServiceImpl) hashAPIKey(key string) (string, error) {
	// For now, use simple hash - in production, use bcrypt or similar
	return fmt.Sprintf("%x", key), nil
}

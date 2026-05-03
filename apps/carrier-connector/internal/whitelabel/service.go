package whitelabel

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides whitelabel management operations
type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewService creates a new whitelabel service
func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// CreateBranding creates a new branding configuration
func (s *Service) CreateBranding(ctx context.Context, config *BrandingConfig) error {
	if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
		s.logger.WithError(err).Error("Failed to create branding config")
		return fmt.Errorf("failed to create branding: %w", err)
	}
	s.logger.WithField("tenant_id", config.TenantID).Info("Branding config created")
	return nil
}

// GetBranding retrieves branding by tenant ID
func (s *Service) GetBranding(ctx context.Context, tenantID string) (*BrandingConfig, error) {
	var config BrandingConfig
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&config).Error; err != nil {
		return nil, fmt.Errorf("branding not found: %w", err)
	}
	return &config, nil
}

// GetBrandingByDomain retrieves branding by custom domain
func (s *Service) GetBrandingByDomain(ctx context.Context, domain string) (*BrandingConfig, error) {
	var config BrandingConfig
	if err := s.db.WithContext(ctx).Where("custom_domain = ? AND is_active = ?", domain, true).First(&config).Error; err != nil {
		return nil, fmt.Errorf("branding not found for domain: %w", err)
	}
	return &config, nil
}

// UpdateBranding updates branding configuration
func (s *Service) UpdateBranding(ctx context.Context, config *BrandingConfig) error {
	if err := s.db.WithContext(ctx).Save(config).Error; err != nil {
		return fmt.Errorf("failed to update branding: %w", err)
	}
	return nil
}

// CreatePartnerConfig creates partner configuration
func (s *Service) CreatePartnerConfig(ctx context.Context, config *PartnerConfig) error {
	if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create partner config: %w", err)
	}
	return nil
}

// GetPartnerConfig retrieves partner configuration
func (s *Service) GetPartnerConfig(ctx context.Context, tenantID string) (*PartnerConfig, error) {
	var config PartnerConfig
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&config).Error; err != nil {
		return nil, fmt.Errorf("partner config not found: %w", err)
	}
	return &config, nil
}

// CreateEmailTemplate creates a custom email template
func (s *Service) CreateEmailTemplate(ctx context.Context, template *EmailTemplate) error {
	if err := s.db.WithContext(ctx).Create(template).Error; err != nil {
		return fmt.Errorf("failed to create email template: %w", err)
	}
	return nil
}

// GetEmailTemplate retrieves an email template
func (s *Service) GetEmailTemplate(ctx context.Context, tenantID, templateKey string) (*EmailTemplate, error) {
	var template EmailTemplate
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND template_key = ? AND is_active = ?", tenantID, templateKey, true).First(&template).Error; err != nil {
		return nil, fmt.Errorf("email template not found: %w", err)
	}
	return &template, nil
}

// ListEmailTemplates lists all templates for a tenant
func (s *Service) ListEmailTemplates(ctx context.Context, tenantID string) ([]*EmailTemplate, error) {
	var templates []*EmailTemplate
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

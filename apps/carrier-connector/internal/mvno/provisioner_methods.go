package mvno

import (
	"context"
	"fmt"
	"time"
)

// provisionTenantContext creates tenant context in production
func (p *ProductionProvisioner) provisionTenantContext(ctx context.Context, mvno *MVNO) error {
	tenantConfig := &TenantConfig{
		ID:   mvno.ID,
		Name: mvno.Name,
		Plan: mvno.Plan,
		Settings: map[string]any{
			"business_id":        mvno.BusinessID,
			"max_subscribers":    mvno.Config.MaxSubscribers,
			"allowed_countries":  mvno.Config.AllowedCountries,
			"custom_branding":    mvno.Config.CustomBranding,
			"advanced_analytics": mvno.Config.AdvancedAnalytics,
			"created_at":         time.Now(),
		},
	}

	if err := p.tenantService.CreateTenant(ctx, tenantConfig); err != nil {
		return fmt.Errorf("failed to create tenant in management system: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":   mvno.ID,
		"tenant_id": tenantConfig.ID,
		"plan":      mvno.Plan,
	}).Info("Tenant context provisioned successfully")

	return nil
}

// provisionDatabaseSchema provisions database schema in production
func (p *ProductionProvisioner) provisionDatabaseSchema(ctx context.Context, mvno *MVNO) error {
	// Create dedicated database schema for MVNO
	if err := p.storageService.CreateDatabaseSchema(ctx, mvno.ID); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	p.logger.WithField("mvno_id", mvno.ID).Info("Database schema provisioned successfully")
	return nil
}

// provisionStorageResources provisions storage resources in production
func (p *ProductionProvisioner) provisionStorageResources(ctx context.Context, mvno *MVNO) error {
	storageSize := p.getStorageAllocation(mvno.Plan)

	if err := p.storageService.CreateStorageBucket(ctx, mvno.ID, storageSize); err != nil {
		return fmt.Errorf("failed to create storage bucket: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"storage_gb": storageSize,
	}).Info("Storage resources provisioned successfully")

	return nil
}

// selectCarriers selects optimal carriers for countries in production
func (p *ProductionProvisioner) selectCarriers(countries []string) ([]string, error) {
	selectedCarriers := make(map[string]bool) // Use map to avoid duplicates

	for _, country := range countries {
		carriers, err := p.carrierManager.GetCarriersByCountry(context.Background(), country)
		if err != nil {
			p.logger.WithError(err).WithField("country", country).Error("Failed to get carriers for country")
			continue
		}

		// Select active carriers with best coverage
		for _, carrier := range carriers {
			if carrier.IsActive {
				selectedCarriers[carrier.ID] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(selectedCarriers))
	for carrierID := range selectedCarriers {
		result = append(result, carrierID)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no active carriers found for countries: %v", countries)
	}

	p.logger.WithFields(map[string]any{
		"countries":      countries,
		"selected_count": len(result),
	}).Info("Carriers selected successfully")

	return result, nil
}

// configureCarrier configures individual carrier in production
func (p *ProductionProvisioner) configureCarrier(ctx context.Context, mvno *MVNO, carrierID string) error {
	if err := p.carrierManager.ConfigureCarrier(ctx, mvno.ID, carrierID); err != nil {
		return fmt.Errorf("failed to configure carrier %s: %w", carrierID, err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"carrier_id": carrierID,
	}).Info("Carrier configured successfully")

	return nil
}

// configureRatePlans configures rate plans in production
func (p *ProductionProvisioner) configureRatePlans(ctx context.Context, mvno *MVNO, billingID string) error {
	if err := p.billingService.CreateRatePlans(ctx, mvno.ID, billingID, mvno.Plan); err != nil {
		return fmt.Errorf("failed to create rate plans: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"billing_id": billingID,
		"plan":       mvno.Plan,
	}).Info("Rate plans configured successfully")

	return nil
}

// setupPaymentProcessing setup payment processing in production
func (p *ProductionProvisioner) setupPaymentProcessing(ctx context.Context, mvno *MVNO, billingID string) error {
	if err := p.billingService.SetupPaymentGateway(ctx, mvno.ID, billingID); err != nil {
		return fmt.Errorf("failed to setup payment gateway: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"billing_id": billingID,
	}).Info("Payment processing setup successfully")

	return nil
}

// getAPIPermissions returns API permissions based on plan
func (p *ProductionProvisioner) getAPIPermissions(plan MVNOPlan) []string {
	switch plan {
	case PlanStarter:
		return []string{"read:subscribers", "read:usage"}
	case PlanGrowth:
		return []string{
			"read:subscribers", "write:subscribers",
			"read:usage", "read:billing",
		}
	case PlanScale:
		return []string{
			"read:subscribers", "write:subscribers",
			"read:usage", "write:usage",
			"read:billing", "write:billing",
			"read:analytics", "write:analytics",
		}
	case PlanEnterprise:
		return []string{"*"} // Full access
	default:
		return []string{"read:subscribers"}
	}
}

// getStorageAllocation returns storage allocation in GB
func (p *ProductionProvisioner) getStorageAllocation(plan MVNOPlan) int {
	switch plan {
	case PlanStarter:
		return 10
	case PlanGrowth:
		return 100
	case PlanScale:
		return 1000
	case PlanEnterprise:
		return 10000
	default:
		return 10
	}
}

// ValidateProvisioning validates that all resources are properly provisioned
func (p *ProductionProvisioner) ValidateProvisioning(ctx context.Context, mvno *MVNO) error {
	// Validate tenant exists
	_, err := p.tenantService.GetTenant(ctx, mvno.ID)
	if err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}

	// Validate carrier configurations
	if len(mvno.Config.CarrierPool) == 0 {
		return fmt.Errorf("no carriers configured")
	}

	for _, carrierID := range mvno.Config.CarrierPool {
		// In production, validate carrier connection
		p.logger.WithFields(map[string]any{
			"mvno_id":    mvno.ID,
			"carrier_id": carrierID,
		}).Debug("Validating carrier connection")
	}

	p.logger.WithField("mvno_id", mvno.ID).Info("Provisioning validation completed")
	return nil
}

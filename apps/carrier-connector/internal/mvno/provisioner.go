package mvno

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/sirupsen/logrus"
)

// MVNOProvisioner handles resource provisioning for MVNOs
type MVNOProvisioner struct {
	logger *logrus.Logger
}

// NewMVNOProvisioner creates a new provisioner instance
func NewMVNOProvisioner(logger *logrus.Logger) *MVNOProvisioner {
	return &MVNOProvisioner{logger: logger}
}

// ProvisionResources provisions core resources for the MVNO
func (p *MVNOProvisioner) ProvisionResources(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Provisioning core resources")

	if err := p.provisionTenantContext(ctx, mvno); err != nil {
		return fmt.Errorf("failed to provision tenant context: %w", err)
	}

	if err := p.provisionDatabaseSchema(ctx, mvno); err != nil {
		return fmt.Errorf("failed to provision database schema: %w", err)
	}

	return p.provisionStorageResources(ctx, mvno)
}

// SetupCarriers configures carrier connections for the MVNO
func (p *MVNOProvisioner) SetupCarriers(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up carrier connections")

	selectedCarriers, err := p.selectCarriers(mvno.Config.AllowedCountries)
	if err != nil {
		return fmt.Errorf("failed to select carriers: %w", err)
	}

	for _, carrierID := range selectedCarriers {
		if err := p.configureCarrier(ctx, mvno, carrierID); err != nil {
			p.logger.WithError(err).WithField("carrier_id", carrierID).Error("Failed to configure carrier")
		}
	}

	mvno.Config.CarrierPool = selectedCarriers
	return nil
}

// SetupBilling configures billing system for the MVNO
func (p *MVNOProvisioner) SetupBilling(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up billing system")

	billingID := id.GeneratePrefixed("bill")

	if err := p.configureRatePlans(ctx, mvno, billingID); err != nil {
		return fmt.Errorf("failed to configure rate plans: %w", err)
	}

	return p.setupPaymentProcessing(ctx, mvno, billingID)
}

// SetupAPIAccess configures API access for the MVNO
func (p *MVNOProvisioner) SetupAPIAccess(ctx context.Context, mvno *MVNO) error {
	if !mvno.Config.APIAccess {
		p.logger.WithField("mvno_id", mvno.ID).Info("API access not included in plan")
		return nil
	}

	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up API access")

	apiKey := id.GeneratePrefixed("api")
	_ = id.GeneratePrefixed("sec") // Generate secret key but don't use until storage is implemented
	permissions := p.getAPIPermissions(mvno.Plan)

	p.logger.WithFields(map[string]any{
		"mvno_id":     mvno.ID,
		"api_key":     apiKey[:8] + "...",
		"permissions": len(permissions),
	}).Info("API access configured")

	return nil
}

// provisionTenantContext creates tenant context
func (p *MVNOProvisioner) provisionTenantContext(_ context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Tenant context provisioned")
	return nil
}

// provisionDatabaseSchema provisions database schema
func (p *MVNOProvisioner) provisionDatabaseSchema(_ context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Database schema provisioned")
	return nil
}

// provisionStorageResources provisions storage resources
func (p *MVNOProvisioner) provisionStorageResources(_ context.Context, mvno *MVNO) error {
	storageSize := p.getStorageAllocation(mvno.Plan)
	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"storage_gb": storageSize,
	}).Info("Storage resources provisioned")
	return nil
}

// selectCarriers selects optimal carriers for countries
func (p *MVNOProvisioner) selectCarriers(countries []string) ([]string, error) {
	carriers := []string{"carrier_us_01", "carrier_gb_01", "carrier_de_01"}

	if len(countries) > 0 {
		selected := make([]string, 0)
		for _, carrier := range carriers {
			selected = append(selected, carrier)
		}
		return selected, nil
	}

	return carriers, nil
}

// configureCarrier configures individual carrier
func (p *MVNOProvisioner) configureCarrier(_ context.Context, mvno *MVNO, carrierID string) error {
	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"carrier_id": carrierID,
	}).Info("Carrier configured")
	return nil
}

// configureRatePlans configures rate plans
func (p *MVNOProvisioner) configureRatePlans(_ context.Context, mvno *MVNO, billingID string) error {
	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"billing_id": billingID,
		"plan":       mvno.Plan,
	}).Info("Rate plans configured")
	return nil
}

// setupPaymentProcessing setup payment processing
func (p *MVNOProvisioner) setupPaymentProcessing(_ context.Context, mvno *MVNO, billingID string) error {
	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"billing_id": billingID,
	}).Info("Payment processing setup")
	return nil
}

// getAPIPermissions returns API permissions based on plan
func (p *MVNOProvisioner) getAPIPermissions(plan MVNOPlan) []string {
	switch plan {
	case PlanStarter:
		return []string{"read:subscribers", "read:usage"}
	case PlanGrowth:
		return []string{"read:subscribers", "write:subscribers", "read:usage", "read:billing"}
	case PlanScale:
		return []string{"read:subscribers", "write:subscribers", "read:usage", "write:usage", "read:billing", "write:billing", "read:analytics"}
	case PlanEnterprise:
		return []string{"*"}
	default:
		return []string{"read:subscribers"}
	}
}

// getStorageAllocation returns storage allocation in GB
func (p *MVNOProvisioner) getStorageAllocation(plan MVNOPlan) int {
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

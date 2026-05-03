package mvno

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/sirupsen/logrus"
)

// ProductionProvisioner implements production-ready MVNO provisioning
type ProductionProvisioner struct {
	logger         *logrus.Logger
	tenantService  TenantService
	carrierManager CarrierManager
	billingService BillingService
	storageService StorageService
}

// TenantService interface for tenant management
type TenantService interface {
	CreateTenant(ctx context.Context, tenant *TenantConfig) error
	GetTenant(ctx context.Context, tenantID string) (*TenantConfig, error)
}

// CarrierManager interface for carrier operations
type CarrierManager interface {
	GetCarriersByCountry(ctx context.Context, country string) ([]CarrierInfo, error)
	ConfigureCarrier(ctx context.Context, mvnoID, carrierID string) error
}

// BillingService interface for billing operations
type BillingService interface {
	CreateBillingAccount(ctx context.Context, mvnoID string, plan MVNOPlan) (string, error)
	CreateRatePlans(ctx context.Context, mvnoID, billingID string, plan MVNOPlan) error
	SetupPaymentGateway(ctx context.Context, mvnoID, billingID string) error
}

// StorageService interface for storage operations
type StorageService interface {
	CreateStorageBucket(ctx context.Context, mvnoID string, sizeGB int) error
	CreateDatabaseSchema(ctx context.Context, mvnoID string) error
}

// TenantConfig represents tenant configuration
type TenantConfig struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Plan     MVNOPlan       `json:"plan"`
	Settings map[string]any `json:"settings"`
}

// CarrierInfo represents carrier information
type CarrierInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Countries   []string `json:"countries"`
	SMDEndpoint string   `json:"smd_endpoint"`
	IsActive    bool     `json:"is_active"`
}

// NewProductionProvisioner creates a new production provisioner
func NewProductionProvisioner(
	logger *logrus.Logger,
	tenantService TenantService,
	carrierManager CarrierManager,
	billingService BillingService,
	storageService StorageService,
) *ProductionProvisioner {
	return &ProductionProvisioner{
		logger:         logger,
		tenantService:  tenantService,
		carrierManager: carrierManager,
		billingService: billingService,
		storageService: storageService,
	}
}

// ProvisionResources provisions core resources for the MVNO
func (p *ProductionProvisioner) ProvisionResources(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Provisioning production resources")

	// Create tenant context
	if err := p.provisionTenantContext(ctx, mvno); err != nil {
		return fmt.Errorf("failed to provision tenant context: %w", err)
	}

	// Provision database schema
	if err := p.provisionDatabaseSchema(ctx, mvno); err != nil {
		return fmt.Errorf("failed to provision database schema: %w", err)
	}

	// Provision storage resources
	if err := p.provisionStorageResources(ctx, mvno); err != nil {
		return fmt.Errorf("failed to provision storage resources: %w", err)
	}

	return nil
}

// SetupCarriers configures carrier connections for the MVNO
func (p *ProductionProvisioner) SetupCarriers(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up production carrier connections")

	selectedCarriers, err := p.selectCarriers(mvno.Config.AllowedCountries)
	if err != nil {
		return fmt.Errorf("failed to select carriers: %w", err)
	}

	// Configure each carrier
	for _, carrierID := range selectedCarriers {
		if err := p.configureCarrier(ctx, mvno, carrierID); err != nil {
			p.logger.WithError(err).WithField("carrier_id", carrierID).Error("Failed to configure carrier")
			return fmt.Errorf("failed to configure carrier %s: %w", carrierID, err)
		}
	}

	mvno.Config.CarrierPool = selectedCarriers
	p.logger.WithFields(map[string]any{
		"mvno_id":       mvno.ID,
		"carrier_count": len(selectedCarriers),
	}).Info("Carriers configured successfully")

	return nil
}

// SetupBilling configures billing system for the MVNO
func (p *ProductionProvisioner) SetupBilling(ctx context.Context, mvno *MVNO) error {
	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up production billing system")

	// Create billing account
	billingID, err := p.billingService.CreateBillingAccount(ctx, mvno.ID, mvno.Plan)
	if err != nil {
		return fmt.Errorf("failed to create billing account: %w", err)
	}

	// Configure rate plans
	if err := p.configureRatePlans(ctx, mvno, billingID); err != nil {
		return fmt.Errorf("failed to configure rate plans: %w", err)
	}

	// Setup payment processing
	if err := p.setupPaymentProcessing(ctx, mvno, billingID); err != nil {
		return fmt.Errorf("failed to setup payment processing: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":    mvno.ID,
		"billing_id": billingID,
		"plan":       mvno.Plan,
	}).Info("Billing system configured successfully")

	return nil
}

// SetupAPIAccess configures API access for the MVNO
func (p *ProductionProvisioner) SetupAPIAccess(ctx context.Context, mvno *MVNO) error {
	if !mvno.Config.APIAccess {
		p.logger.WithField("mvno_id", mvno.ID).Info("API access not included in plan")
		return nil
	}

	p.logger.WithField("mvno_id", mvno.ID).Info("Setting up production API access")

	// Generate API credentials
	apiKey := id.GeneratePrefixed("api")
	secretKey := id.GeneratePrefixed("sec")
	permissions := p.getAPIPermissions(mvno.Plan)

	// Store API configuration in tenant service
	tenant, err := p.tenantService.GetTenant(ctx, mvno.ID)
	if err != nil {
		return fmt.Errorf("failed to get tenant for API setup: %w", err)
	}

	tenant.Settings["api_access"] = map[string]any{
		"api_key":     apiKey,
		"secret_key":  secretKey,
		"permissions": permissions,
		"created_at":  time.Now(),
		"expires_at":  time.Now().AddDate(1, 0, 0), // 1 year expiry
	}

	if err := p.tenantService.CreateTenant(ctx, tenant); err != nil {
		return fmt.Errorf("failed to store API configuration: %w", err)
	}

	p.logger.WithFields(map[string]any{
		"mvno_id":     mvno.ID,
		"api_key":     apiKey[:8] + "...",
		"permissions": len(permissions),
	}).Info("API access configured successfully")

	return nil
}

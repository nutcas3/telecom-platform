package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
)

// SMDPService provides high-level SM-DP+ operations with multi-carrier support
type SMDPService struct {
	manager    *smdp.SMDPManager
	repository *repository.PostgresProfileStore
	logger     *logrus.Logger
}

// NewSMDPService creates a new SM-DP+ service
func NewSMDPService(repo *repository.PostgresProfileStore) *SMDPService {
	config := smdp.DefaultManagerConfig()
	manager := smdp.NewSMDPManager(repo, config)

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Start health checking
	ctx := context.Background()
	go manager.StartHealthChecking(ctx)

	return &SMDPService{
		manager:    manager,
		repository: repo,
		logger:     logger,
	}
}

// InitializeDefaultCarriers adds common carriers to the manager
func (s *SMDPService) InitializeDefaultCarriers() error {
	defaultCarriers := []*smdp.CarrierConfig{
		{
			ID:                    "carrier-att-us",
			Name:                  "AT&T US",
			CountryCode:           "US",
			MCC:                   "310",
			MNC:                   "410",
			ES2BaseURL:            "https://smdp.att.com",
			ES2APIKey:             "",
			ES2InsecureSkip:       false,
			Priority:              100,
			IsActive:              true,
			MaxConcurrentReqs:     100,
			SupportedProfileTypes: []string{"operational", "test"},
			SupportedMCCs:         []string{"310", "311", "312", "313", "314", "315", "316"},
			Features:              []string{"bulk_download", "remote_provisioning"},
		},
		{
			ID:                    "carrier-verizon-us",
			Name:                  "Verizon US",
			CountryCode:           "US",
			MCC:                   "311",
			MNC:                   "480",
			ES2BaseURL:            "https://smdp.verizon.com",
			ES2APIKey:             "",
			ES2InsecureSkip:       false,
			Priority:              90,
			IsActive:              true,
			MaxConcurrentReqs:     80,
			SupportedProfileTypes: []string{"operational", "test"},
			SupportedMCCs:         []string{"311", "312"},
			Features:              []string{"bulk_download"},
		},
		{
			ID:                    "carrier-tmobile-de",
			Name:                  "T-Mobile DE",
			CountryCode:           "DE",
			MCC:                   "262",
			MNC:                   "01",
			ES2BaseURL:            "https://smdp.t-mobile.de",
			ES2APIKey:             "",
			ES2InsecureSkip:       false,
			Priority:              85,
			IsActive:              true,
			MaxConcurrentReqs:     60,
			SupportedProfileTypes: []string{"operational"},
			SupportedMCCs:         []string{"262"},
			Features:              []string{"bulk_download", "remote_provisioning"},
		},
	}

	for _, carrierConfig := range defaultCarriers {
		carrier := carrierConfig.ToCarrier()
		if err := s.manager.AddCarrier(carrier); err != nil {
			s.logger.WithError(err).WithField("carrier_id", carrier.ID).
				Warn("Failed to add default carrier")
			continue
		}
		s.logger.WithField("carrier_id", carrier.ID).Info("Added default carrier")
	}

	return nil
}

// DownloadProfile handles profile download with intelligent carrier selection
func (s *SMDPService) DownloadProfile(ctx context.Context, req *smdp.ProfileRequest) (*smdp.ProfileResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"eid":       req.EID,
		"iccid":     req.ICCID,
		"preferred": req.PreferredCarrier,
	}).Info("Processing profile download through SM-DP+ service")

	return s.manager.DownloadProfile(ctx, req)
}

// GetOptimalCarrier selects the best carrier for a given request
func (s *SMDPService) GetOptimalCarrier(ctx context.Context, req *smdp.ProfileRequest) (*smdp.Carrier, error) {
	return s.manager.SelectCarrier(ctx)
}

// GetCarrierHealth returns health status of all carriers
func (s *SMDPService) GetCarrierHealth() map[string]*smdp.Carrier {
	return s.manager.GetCarrierStatus()
}

// AddCarrier dynamically adds a new carrier
func (s *SMDPService) AddCarrier(carrierConfig *smdp.CarrierConfig) error {
	carrier := carrierConfig.ToCarrier()

	if err := s.manager.AddCarrier(carrier); err != nil {
		return fmt.Errorf("failed to add carrier %s: %w", carrier.ID, err)
	}

	s.logger.WithField("carrier_id", carrier.ID).Info("Successfully added new carrier")
	return nil
}

// RemoveCarrier removes a carrier from the system
func (s *SMDPService) RemoveCarrier(carrierID string) error {
	if err := s.manager.RemoveCarrier(carrierID); err != nil {
		return fmt.Errorf("failed to remove carrier %s: %w", carrierID, err)
	}

	s.logger.WithField("carrier_id", carrierID).Info("Successfully removed carrier")
	return nil
}

// GetManager returns the underlying SMDPManager
func (s *SMDPService) GetManager() *smdp.SMDPManager {
	return s.manager
}

// GetCarrierMetrics returns performance metrics for all carriers
func (s *SMDPService) GetCarrierMetrics() map[string]*smdp.CarrierMetrics {
	carriers := s.manager.GetCarrierStatus()
	metrics := make(map[string]*smdp.CarrierMetrics)

	for id, carrier := range carriers {
		metrics[id] = carrier.Metrics
	}

	return metrics
}

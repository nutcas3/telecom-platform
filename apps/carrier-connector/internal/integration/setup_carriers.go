package integration

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// SetupCarriers loads real carrier configurations from database or config files
func (si *SelectionIntegration) SetupCarriers() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Loading carriers from configuration")

	// Load carriers from configuration file
	configPath := "configs/carriers.json"

	config, err := LoadCarriersFromFile(configPath)
	if err != nil {
		logger.WithError(err).Error("Failed to load carriers from config file")
		return fmt.Errorf("failed to load carriers from config: %w", err)
	}

	carriers, err := ConvertConfigToCarriers(config)
	if err != nil {
		logger.WithError(err).Error("Failed to convert config to carriers")
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Validate and add carriers to SMDP manager
	successCount := 0
	for _, carrier := range carriers {
		if err := validateCarrier(carrier); err != nil {
			logger.WithError(err).WithField("carrier_id", carrier.ID).Error("Carrier validation failed")
			continue
		}

		if err := si.manager.AddCarrier(carrier); err != nil {
			logger.WithError(err).WithField("carrier_id", carrier.ID).Error("Failed to add carrier to manager")
			continue
		}
		successCount++
	}

	logger.WithFields(logrus.Fields{
		"config_file":   configPath,
		"total_loaded":  len(carriers),
		"success_count": successCount,
		"failed_count":  len(carriers) - successCount,
	}).Info("Carriers loaded from configuration file")

	return nil
}

// validateCarrier validates carrier configuration
func validateCarrier(carrier *smdp.Carrier) error {
	if carrier.ID == "" {
		return fmt.Errorf("carrier ID is required")
	}
	if carrier.Name == "" {
		return fmt.Errorf("carrier name is required")
	}
	if carrier.MCC == "" {
		return fmt.Errorf("carrier MCC is required")
	}
	if carrier.MNC == "" {
		return fmt.Errorf("carrier MNC is required")
	}
	if carrier.CountryCode == "" {
		return fmt.Errorf("carrier country code is required")
	}
	if carrier.ES2Config == nil {
		return fmt.Errorf("carrier ES2 config is required")
	}
	if carrier.ES2Config.BaseURL == "" {
		return fmt.Errorf("carrier ES2 base URL is required")
	}
	if carrier.ES2Config.APIKey == "" {
		return fmt.Errorf("carrier ES2 API key is required")
	}

	return nil
}

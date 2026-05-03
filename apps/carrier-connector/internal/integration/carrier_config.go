package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
)

// NewGormCarrierRepository creates a new GORM carrier repository
func NewGormCarrierRepository(db *gorm.DB, logger *logrus.Logger) *GormCarrierRepository {
	return &GormCarrierRepository{
		db:     db,
		logger: logger,
	}
}

// CarrierModel represents the database model for carriers
type CarrierModel struct {
	ID           string    `gorm:"primaryKey;column:id" json:"id"`
	Name         string    `gorm:"column:name" json:"name"`
	MCC          string    `gorm:"column:mcc" json:"mcc"`
	MNC          string    `gorm:"column:mnc" json:"mnc"`
	CountryCode  string    `gorm:"column:country_code" json:"country_code"`
	IsActive     bool      `gorm:"column:is_active" json:"is_active"`
	Priority     int       `gorm:"column:priority" json:"priority"`
	ES2Config    string    `gorm:"column:es2_config;type:text" json:"es2_config"`
	Capabilities string    `gorm:"column:capabilities;type:text" json:"capabilities"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name for the carrier model
func (CarrierModel) TableName() string {
	return "carriers"
}

// GetCarriers retrieves all carriers from the database
func (r *GormCarrierRepository) GetCarriers(ctx context.Context) ([]*smdp.Carrier, error) {
	var models []CarrierModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&models).Error; err != nil {
		r.logger.Error("Failed to get carriers from database", "error", err)
		return nil, fmt.Errorf("failed to get carriers: %w", err)
	}

	carriers := make([]*smdp.Carrier, 0, len(models))
	for _, model := range models {
		carrier, err := r.modelToCarrier(&model)
		if err != nil {
			r.logger.Error("Failed to convert carrier model", "error", err, "carrier_id", model.ID)
			continue
		}
		carriers = append(carriers, carrier)
	}

	return carriers, nil
}

// GetCarrier retrieves a specific carrier by ID
func (r *GormCarrierRepository) GetCarrier(ctx context.Context, id string) (*smdp.Carrier, error) {
	var model CarrierModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("carrier not found: %s", id)
		}
		r.logger.Error("Failed to get carrier from database", "error", err, "carrier_id", id)
		return nil, fmt.Errorf("failed to get carrier: %w", err)
	}

	return r.modelToCarrier(&model)
}

// SaveCarrier saves a carrier to the database
func (r *GormCarrierRepository) SaveCarrier(ctx context.Context, carrier *smdp.Carrier) error {
	model := r.carrierToModel(carrier)

	// Check if carrier exists and update or create accordingly
	var existing CarrierModel
	err := r.db.WithContext(ctx).Where("id = ?", model.ID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new carrier
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			r.logger.WithError(err).Error("Failed to create carrier")
			return fmt.Errorf("failed to create carrier: %w", err)
		}
	} else if err != nil {
		r.logger.WithError(err).Error("Failed to check carrier existence")
		return fmt.Errorf("failed to check carrier: %w", err)
	} else {
		// Update existing carrier
		model.UpdatedAt = time.Now()
		if err := r.db.WithContext(ctx).Model(&existing).Updates(model).Error; err != nil {
			r.logger.WithError(err).Error("Failed to update carrier")
			return fmt.Errorf("failed to update carrier: %w", err)
		}
	}

	return nil
}

// UpdateCarrierMetrics updates carrier metrics
func (r *GormCarrierRepository) UpdateCarrierMetrics(ctx context.Context, id string, metrics *smdp.CarrierMetrics) error {
	// Persist metrics alongside the carrier record
	updates := map[string]any{
		"updated_at": time.Now(),
	}

	if err := r.db.WithContext(ctx).Table("carriers").Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update carrier record timestamp")
		return fmt.Errorf("failed to update carrier metrics: %w", err)
	}

	// Store detailed metrics in a dedicated metrics table
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal carrier metrics: %w", err)
	}

	metricsRecord := struct {
		CarrierID   string    `gorm:"column:carrier_id;index"`
		MetricsData string    `gorm:"column:metrics_data;type:text"`
		RecordedAt  time.Time `gorm:"column:recorded_at"`
	}{
		CarrierID:   id,
		MetricsData: string(metricsJSON),
		RecordedAt:  time.Now(),
	}

	if err := r.db.WithContext(ctx).Table("carrier_metrics").Create(&metricsRecord).Error; err != nil {
		r.logger.WithError(err).Error("Failed to persist carrier metrics")
		return fmt.Errorf("failed to store carrier metrics: %w", err)
	}

	successRate := float64(0)
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
	}
	r.logger.WithField("carrier_id", id).
		WithField("total_requests", metrics.TotalRequests).
		WithField("success_rate", successRate).
		Info("Carrier metrics updated")

	return nil
}

// Helper methods for model conversion

func (r *GormCarrierRepository) modelToCarrier(model *CarrierModel) (*smdp.Carrier, error) {
	carrier := &smdp.Carrier{
		ID:          model.ID,
		Name:        model.Name,
		MCC:         model.MCC,
		MNC:         model.MNC,
		CountryCode: model.CountryCode,
		IsActive:    model.IsActive,
		Priority:    model.Priority,
	}

	// Parse ES2Config
	if model.ES2Config != "" {
		var es2Config config.ES2Config
		if err := json.Unmarshal([]byte(model.ES2Config), &es2Config); err != nil {
			return nil, fmt.Errorf("failed to parse ES2 config: %w", err)
		}
		carrier.ES2Config = &es2Config
	}

	// Parse Capabilities
	if model.Capabilities != "" {
		var capabilities smdp.CarrierCapabilities
		if err := json.Unmarshal([]byte(model.Capabilities), &capabilities); err != nil {
			return nil, fmt.Errorf("failed to parse carrier capabilities: %w", err)
		}
		carrier.Capabilities = &capabilities
	}

	// Initialize metrics (in production, this would come from a metrics table)
	carrier.Metrics = &smdp.CarrierMetrics{
		TotalRequests:       0,
		SuccessfulRequests:  0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		RequestRate:         0,
	}

	return carrier, nil
}

func (r *GormCarrierRepository) carrierToModel(carrier *smdp.Carrier) *CarrierModel {
	model := &CarrierModel{
		ID:          carrier.ID,
		Name:        carrier.Name,
		MCC:         carrier.MCC,
		MNC:         carrier.MNC,
		CountryCode: carrier.CountryCode,
		IsActive:    carrier.IsActive,
		Priority:    carrier.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Serialize ES2Config
	if carrier.ES2Config != nil {
		if data, err := json.Marshal(carrier.ES2Config); err == nil {
			model.ES2Config = string(data)
		}
	}

	// Serialize Capabilities
	if carrier.Capabilities != nil {
		if data, err := json.Marshal(carrier.Capabilities); err == nil {
			model.Capabilities = string(data)
		}
	}

	return model
}

// LoadCarriersFromFile loads carrier configurations from a JSON/YAML file
func LoadCarriersFromFile(configPath string) (*CarrierConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read carrier config file: %w", err)
	}

	var config CarrierConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse carrier config file: %w", err)
	}

	return &config, nil
}

// ConvertConfigToCarriers converts carrier definitions to smdp.Carrier objects
func ConvertConfigToCarriers(config *CarrierConfig) ([]*smdp.Carrier, error) {
	carriers := make([]*smdp.Carrier, 0, len(config.Carriers))

	for _, def := range config.Carriers {
		carrier := &smdp.Carrier{
			ID:           def.ID,
			Name:         def.Name,
			MCC:          def.MCC,
			MNC:          def.MNC,
			CountryCode:  def.CountryCode,
			IsActive:     def.IsActive,
			Priority:     def.Priority,
			ES2Config:    def.ES2Config,
			Capabilities: def.Capabilities,
			Metrics: &smdp.CarrierMetrics{
				TotalRequests:       0,
				SuccessfulRequests:  0,
				FailedRequests:      0,
				AverageResponseTime: 0,
				RequestRate:         0,
			},
		}

		carriers = append(carriers, carrier)
	}

	return carriers, nil
}

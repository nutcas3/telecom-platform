package integration

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CarrierConfig represents the configuration structure for carriers
type CarrierConfig struct {
	Carriers []CarrierDefinition `json:"carriers" yaml:"carriers"`
}

// CarrierDefinition defines a carrier configuration
type CarrierDefinition struct {
	ID           string                    `json:"id" yaml:"id"`
	Name         string                    `json:"name" yaml:"name"`
	MCC          string                    `json:"mcc" yaml:"mcc"`
	MNC          string                    `json:"mnc" yaml:"mnc"`
	CountryCode  string                    `json:"country_code" yaml:"country_code"`
	IsActive     bool                      `json:"is_active" yaml:"is_active"`
	Priority     int                       `json:"priority" yaml:"priority"`
	ES2Config    *config.ES2Config         `json:"es2_config" yaml:"es2_config"`
	Capabilities *smdp.CarrierCapabilities `json:"capabilities" yaml:"capabilities"`
}

// CarrierRepository defines the interface for carrier data operations
type CarrierRepository interface {
	GetCarriers(ctx context.Context) ([]*smdp.Carrier, error)
	GetCarrier(ctx context.Context, id string) (*smdp.Carrier, error)
	SaveCarrier(ctx context.Context, carrier *smdp.Carrier) error
	UpdateCarrierMetrics(ctx context.Context, id string, metrics *smdp.CarrierMetrics) error
}

// GormCarrierRepository implements CarrierRepository using GORM
type GormCarrierRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

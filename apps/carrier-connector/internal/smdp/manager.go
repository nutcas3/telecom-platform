package smdp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/sirupsen/logrus"
)

// SMDPManager manages multiple SM-DP+ carriers
type SMDPManager struct {
	carriers       map[string]*Carrier
	carriersMutex  sync.RWMutex
	es2Clients     map[string]*es2.ES2Client
	clientsMutex   sync.RWMutex
	repository     *repository.PostgresProfileStore
	healthChecker  *HealthChecker
	loadBalancer   *LoadBalancer
	config         *ManagerConfig
	logger         *logrus.Logger
}

// NewSMDPManager creates a new SM-DP+ Manager
func NewSMDPManager(repo *repository.PostgresProfileStore, config *ManagerConfig) *SMDPManager {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &SMDPManager{
		carriers:      make(map[string]*Carrier),
		es2Clients:    make(map[string]*es2.ES2Client),
		repository:    repo,
		config:        config,
		logger:        logger,
		healthChecker: NewHealthChecker(config.HealthCheckInterval),
		loadBalancer:  NewLoadBalancer(),
	}
}

// AddCarrier adds a new carrier to the manager
func (m *SMDPManager) AddCarrier(carrier *Carrier) error {
	m.carriersMutex.Lock()
	defer m.carriersMutex.Unlock()

	if err := m.validateCarrier(carrier); err != nil {
		return fmt.Errorf("invalid carrier configuration: %w", err)
	}

	client := es2.NewES2Client(carrier.ES2Config)
	m.carriers[carrier.ID] = carrier
	m.es2Clients[carrier.ID] = client

	m.logger.WithFields(logrus.Fields{
		"carrier_id":   carrier.ID,
		"carrier_name": carrier.Name,
	}).Info("Added carrier to SM-DP+ Manager")

	return nil
}

// RemoveCarrier removes a carrier from the manager
func (m *SMDPManager) RemoveCarrier(carrierID string) error {
	m.carriersMutex.Lock()
	defer m.carriersMutex.Unlock()

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if _, exists := m.carriers[carrierID]; !exists {
		return fmt.Errorf("carrier %s not found", carrierID)
	}

	delete(m.carriers, carrierID)
	delete(m.es2Clients, carrierID)

	m.logger.WithField("carrier_id", carrierID).Info("Removed carrier from SM-DP+ Manager")
	return nil
}

// GetCarrierStatus returns the current status of all carriers
func (m *SMDPManager) GetCarrierStatus() map[string]*Carrier {
	m.carriersMutex.RLock()
	defer m.carriersMutex.RUnlock()

	status := make(map[string]*Carrier)
	for id, carrier := range m.carriers {
		carrierCopy := *carrier
		status[id] = &carrierCopy
	}
	return status
}

// StartHealthChecking starts the background health checking process
func (m *SMDPManager) StartHealthChecking(ctx context.Context) {
	m.healthChecker.Start(ctx, m.carriers, m.es2Clients, m.updateCarrierHealth)
}

// validateCarrier validates carrier configuration
func (m *SMDPManager) validateCarrier(carrier *Carrier) error {
	if carrier.ID == "" || carrier.Name == "" || carrier.MCC == "" || carrier.MNC == "" {
		return fmt.Errorf("carrier ID, name, MCC, and MNC are required")
	}
	if carrier.ES2Config == nil || carrier.ES2Config.BaseURL == "" {
		return fmt.Errorf("ES2+ configuration and base URL are required")
	}
	return nil
}

// updateCarrierHealth updates the health status of a carrier
func (m *SMDPManager) updateCarrierHealth(carrierID string, status CarrierHealthStatus) {
	m.carriersMutex.Lock()
	defer m.carriersMutex.Unlock()

	if carrier, exists := m.carriers[carrierID]; exists {
		carrier.HealthStatus = status
		carrier.LastHealthCheck = time.Now()
	}
}

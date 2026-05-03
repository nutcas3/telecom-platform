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
	carriers      map[string]*Carrier
	carriersMutex sync.RWMutex
	es2Clients    map[string]*es2.ES2Client
	clientsMutex  sync.RWMutex
	repository    *repository.PostgresProfileStore
	healthChecker *HealthChecker
	loadBalancer  *LoadBalancer
	selector      *SelectionAlgorithm
	config        *ManagerConfig
	logger        *logrus.Logger
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
		selector:      NewSelectionAlgorithm(),
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

// SelectOptimalCarrier selects the best carrier based on criteria
func (m *SMDPManager) SelectOptimalCarrier(ctx context.Context, criteria *SelectionCriteria) (*CarrierScore, error) {
	m.carriersMutex.RLock()
	defer m.carriersMutex.RUnlock()

	// Get all active and healthy carriers
	healthyCarriers := make([]*Carrier, 0)
	for _, carrier := range m.carriers {
		if carrier.IsActive && carrier.HealthStatus == CarrierStatusHealthy {
			healthyCarriers = append(healthyCarriers, carrier)
		}
	}

	return m.selector.SelectOptimalCarrier(ctx, healthyCarriers, criteria)
}

// SelectCarrier selects a carrier using the optimal algorithm with default criteria
func (m *SMDPManager) SelectCarrier(ctx context.Context) (*Carrier, error) {
	criteria := &SelectionCriteria{
		Region:            "",
		ProfileType:       "operational",
		Urgency:           "medium",
		CostSensitivity:   0.5,
		PerformanceWeight: 0.4,
		ReliabilityWeight: 0.4,
	}

	score, err := m.SelectOptimalCarrier(ctx, criteria)
	if err != nil {
		return nil, err
	}

	return score.Carrier, nil
}

// GetSelectionHistory returns the selection history for a carrier
func (m *SMDPManager) GetSelectionHistory(carrierID string) []CarrierScore {
	return m.selector.GetSelectionHistory(carrierID)
}

// UpdateLearning updates the selection algorithm with performance feedback
func (m *SMDPManager) UpdateLearning(carrierID string, actualPerformance float64) {
	m.selector.UpdateLearning(carrierID, actualPerformance)
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

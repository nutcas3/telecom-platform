package smdp

import (
	"context"
	"sync"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/sirupsen/logrus"
)

// HealthChecker performs periodic health checks on carriers
type HealthChecker struct {
	interval       time.Duration
	logger         *logrus.Logger
	healthUpdateFn func(carrierID string, status CarrierHealthStatus)
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		logger:   logrus.New(),
	}
}

// Start begins the health checking process
func (h *HealthChecker) Start(ctx context.Context, carriers map[string]*Carrier, clients map[string]*es2.ES2Client, updateFn func(string, CarrierHealthStatus)) {
	h.healthUpdateFn = updateFn

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Health checker stopped")
			return
		case <-ticker.C:
			h.checkAllCarriers(carriers, clients)
		}
	}
}

// checkAllCarriers performs health checks on all carriers
func (h *HealthChecker) checkAllCarriers(carriers map[string]*Carrier, clients map[string]*es2.ES2Client) {
	var wg sync.WaitGroup

	for carrierID, carrier := range carriers {
		if !carrier.IsActive {
			continue
		}

		wg.Add(1)
		go func(id string, c *Carrier) {
			defer wg.Done()
			h.checkCarrier(id, c, clients[id])
		}(carrierID, carrier)
	}

	wg.Wait()
}

// checkCarrier performs a health check on a single carrier
func (h *HealthChecker) checkCarrier(carrierID string, carrier *Carrier, client *es2.ES2Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	
	// Perform a simple health check by attempting to get profile status
	// This is a lightweight operation that tests connectivity
	req := &es2.GetProfileStatusRequest{
		EID:   "health-check-eid",
		ICCID: "health-check-iccid",
	}

	_, err := client.GetProfileStatus(ctx, req)
	responseTime := time.Since(startTime)

	status := h.evaluateHealth(err, responseTime, carrier.Metrics)
	h.healthUpdateFn(carrierID, status)

	h.logger.WithFields(logrus.Fields{
		"carrier_id":    carrierID,
		"status":        status,
		"response_time": responseTime,
		"error":         err,
	}).Debug("Health check completed")
}

// evaluateHealth evaluates the health status based on error and response time
func (h *HealthChecker) evaluateHealth(err error, responseTime time.Duration, metrics *CarrierMetrics) CarrierHealthStatus {
	if err != nil {
		// Check if this is the first failure or consecutive failures
		if metrics.FailedRequests > 5 {
			return CarrierStatusUnhealthy
		}
		return CarrierStatusDegraded
	}

	// Check response time
	if responseTime > 5*time.Second {
		return CarrierStatusDegraded
	}

	// Check error rate
	if metrics.TotalRequests > 0 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		if errorRate > 0.1 { // More than 10% error rate
			return CarrierStatusDegraded
		}
	}

	return CarrierStatusHealthy
}

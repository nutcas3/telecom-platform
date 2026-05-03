package smdp

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/sirupsen/logrus"
)

func (m *SMDPManager) DownloadProfile(ctx context.Context, req *ProfileRequest) (*ProfileResponse, error) {
	startTime := time.Now()

	m.logger.WithFields(logrus.Fields{
		"eid":       req.EID,
		"preferred": req.PreferredCarrier,
	}).Info("Processing profile download request")

	selectedCarrier, err := m.SelectCarrier(ctx)
	if err != nil {
		return &ProfileResponse{
			Success:       false,
			StatusMessage: fmt.Sprintf("Carrier selection failed: %v", err),
			ResponseTime:  time.Since(startTime),
		}, err
	}

	response, err := m.downloadWithRetry(ctx, req, selectedCarrier, 0)
	if err != nil && m.config.EnableFailover {
		m.logger.WithError(err).Warn("Primary carrier failed, attempting failover")
		return m.handleFailover(ctx, req, selectedCarrier.ID, startTime)
	}

	response.ResponseTime = time.Since(startTime)
	return response, err
}

func (m *SMDPManager) downloadWithRetry(ctx context.Context, req *ProfileRequest, carrier *Carrier, attempt int) (*ProfileResponse, error) {
	m.clientsMutex.RLock()
	client, exists := m.es2Clients[carrier.ID]
	m.clientsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("ES2+ client not found for carrier %s", carrier.ID)
	}

	downloadReq := &es2.DownloadProfileRequest{
		EID:              req.EID,
		ICCID:            req.ICCID,
		ProfileType:      req.ProfileType,
		ConfirmationCode: req.ConfirmationCode,
	}

	startTime := time.Now()
	resp, err := client.DownloadProfile(ctx, downloadReq)
	responseTime := time.Since(startTime)

	m.updateCarrierMetrics(carrier.ID, err == nil, responseTime, err)

	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"carrier_id": carrier.ID,
			"attempt":    attempt + 1,
			"error":      err,
		}).Error("Profile download failed")

		if attempt < m.config.MaxRetries {
			time.Sleep(m.config.RetryDelay)
			return m.downloadWithRetry(ctx, req, carrier, attempt+1)
		}

		return &ProfileResponse{
			Success:       false,
			CarrierID:     carrier.ID,
			StatusMessage: fmt.Sprintf("Download failed after %d attempts: %v", attempt+1, err),
			ResponseTime:  responseTime,
		}, err
	}

	m.logger.WithFields(logrus.Fields{
		"carrier_id":       carrier.ID,
		"execution_status": resp.ExecutionStatus,
		"response_time":    responseTime,
	}).Info("Profile download successful")

	return &ProfileResponse{
		Success:         true,
		CarrierID:       carrier.ID,
		ExecutionStatus: resp.ExecutionStatus,
		StatusMessage:   resp.StatusMessage,
		ResponseTime:    responseTime,
	}, nil
}

// handleFailover attempts to download profile using alternative carriers
func (m *SMDPManager) handleFailover(ctx context.Context, req *ProfileRequest, failedCarrierID string, startTime time.Time) (*ProfileResponse, error) {
	m.carriersMutex.RLock()
	defer m.carriersMutex.RUnlock()

	var retriedOn []string

	for carrierID, carrier := range m.carriers {
		if carrierID == failedCarrierID || !carrier.IsActive || carrier.HealthStatus != CarrierStatusHealthy {
			continue
		}

		m.logger.WithField("carrier_id", carrierID).Info("Attempting failover to carrier")

		response, err := m.downloadWithRetry(ctx, req, carrier, 0)
		if err == nil && response.Success {
			response.RetriedOn = retriedOn
			response.ResponseTime = time.Since(startTime)
			return response, nil
		}

		retriedOn = append(retriedOn, carrierID)
	}

	return &ProfileResponse{
		Success:       false,
		StatusMessage: "All carriers failed during failover",
		ResponseTime:  time.Since(startTime),
		RetriedOn:     retriedOn,
	}, fmt.Errorf("failover exhausted all carriers")
}

// updateCarrierMetrics updates the metrics for a carrier
func (m *SMDPManager) updateCarrierMetrics(carrierID string, success bool, responseTime time.Duration, err error) {
	m.carriersMutex.Lock()
	defer m.carriersMutex.Unlock()

	carrier, exists := m.carriers[carrierID]
	if !exists {
		return
	}

	carrier.Metrics.TotalRequests++
	if success {
		carrier.Metrics.SuccessfulRequests++
	} else {
		carrier.Metrics.FailedRequests++
		carrier.Metrics.LastError = err.Error()
		carrier.Metrics.LastErrorTime = time.Now()
	}

	if carrier.Metrics.TotalRequests == 1 {
		carrier.Metrics.AverageResponseTime = responseTime
	} else {
		carrier.Metrics.AverageResponseTime = time.Duration(
			(int64(carrier.Metrics.AverageResponseTime)*int64(carrier.Metrics.TotalRequests-1) + int64(responseTime)) / int64(carrier.Metrics.TotalRequests),
		)
	}

	carrier.Metrics.RequestRate = float64(carrier.Metrics.TotalRequests) / time.Since(time.Now().Add(-time.Minute)).Seconds()
}

package telecom

import (
	"context"
)

// SystemAPI handles system-related API calls
type SystemAPI struct {
	client *HTTPClient
}

// NewSystemAPI creates a new SystemAPI
func NewSystemAPI(client *HTTPClient) *SystemAPI {
	return &SystemAPI{client: client}
}

// GetStats retrieves system statistics
func (s *SystemAPI) GetStats(ctx context.Context) (*SystemStats, error) {
	var stats SystemStats
	err := s.client.Get(ctx, "/v1/system/stats", &stats)
	return &stats, err
}

// GetHealth retrieves system health status
func (s *SystemAPI) GetHealth(ctx context.Context) (*HealthStatus, error) {
	var health HealthStatus
	err := s.client.Get(ctx, "/v1/health", &health)
	return &health, err
}

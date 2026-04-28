package api

import (
	"fmt"
	"time"
)

// Alert represents a monitoring alert
type Alert struct {
	ID       string    `json:"id"`
	Severity string    `json:"severity"`
	Service  string    `json:"service"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
}

// ListAlerts returns alerts
func (c *Client) ListAlerts() ([]Alert, error) {
	var resp struct {
		Data []Alert `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/alerts", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SystemStats represents system health/metrics
type SystemStats struct {
	ActiveSessions   int     `json:"active_sessions"`
	TotalAccounts    int     `json:"total_accounts"`
	BlockedUsers     int     `json:"blocked_users"`
	LowBalanceAlerts int     `json:"low_balance_alerts"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
	Uptime           string  `json:"uptime"`
}

// GetSystemStats retrieves system stats
func (c *Client) GetSystemStats() (*SystemStats, error) {
	var stats SystemStats
	if err := c.doGetJSON("/api/v1/system/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// HealthStatus represents a health check result
type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	LastCheck time.Time `json:"last_check"`
}

// GetHealth returns per-service health
func (c *Client) GetHealth() ([]HealthStatus, error) {
	var resp struct {
		Data []HealthStatus `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/health", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ServiceHealth represents the health status of a single service
type ServiceHealth struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
	Message  string `json:"message,omitempty"`
}

// CheckAllServices checks the health of all platform services
func (c *Client) CheckAllServices() ([]ServiceHealth, error) {
	services := []struct {
		name     string
		endpoint string
	}{
		{"api-server", "/health"},
		{"charging-engine", "/health"},
		{"carrier-connector", "/api/v1/health"},
		{"packet-gateway", "/health"},
		{"web-dashboard", "/health"},
	}

	healthResults := make([]ServiceHealth, 0, len(services))
	for _, svc := range services {
		health := ServiceHealth{
			Name:     svc.name,
			Endpoint: svc.endpoint,
		}

		resp, err := c.httpClient.Get(c.baseURL + svc.endpoint)
		if err != nil {
			health.Status = "unreachable"
			health.Message = err.Error()
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				health.Status = "healthy"
			} else {
				health.Status = "unhealthy"
				health.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		}

		healthResults = append(healthResults, health)
	}

	return healthResults, nil
}

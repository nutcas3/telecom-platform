package smdp

import (
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
)

// Carrier represents a mobile network operator with SM-DP+ capabilities
type Carrier struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	CountryCode     string               `json:"country_code"`
	MCC             string               `json:"mcc"`
	MNC             string               `json:"mnc"`
	ES2Config       *config.ES2Config    `json:"es2_config"`
	Priority        int                  `json:"priority"` // Higher number = higher priority
	IsActive        bool                 `json:"is_active"`
	HealthStatus    CarrierHealthStatus  `json:"health_status"`
	LastHealthCheck time.Time            `json:"last_health_check"`
	Metrics         *CarrierMetrics      `json:"metrics"`
	Capabilities    *CarrierCapabilities `json:"capabilities"`
}

// CarrierHealthStatus represents the health status of a carrier
type CarrierHealthStatus string

const (
	CarrierStatusHealthy   CarrierHealthStatus = "healthy"
	CarrierStatusDegraded  CarrierHealthStatus = "degraded"
	CarrierStatusUnhealthy CarrierHealthStatus = "unhealthy"
	CarrierStatusUnknown   CarrierHealthStatus = "unknown"
)

// CarrierCapabilities represents the capabilities of a carrier
type CarrierCapabilities struct {
	SupportedProfileTypes []string `json:"supported_profile_types"`
	MaxConcurrentRequests int      `json:"max_concurrent_requests"`
	SupportedMCCs         []string `json:"supported_mccs"`
	SupportedMNCs         []string `json:"supported_mncs"`
	Features              []string `json:"features"` // e.g., ["bulk_download", "remote_provisioning"]
}

// CarrierMetrics tracks performance metrics for a carrier
type CarrierMetrics struct {
	TotalRequests       uint64        `json:"total_requests"`
	SuccessfulRequests  uint64        `json:"successful_requests"`
	FailedRequests      uint64        `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastError           string        `json:"last_error"`
	LastErrorTime       time.Time     `json:"last_error_time"`
	RequestRate         float64       `json:"request_rate"` // requests per second
}

// ProfileRequest represents an eSIM profile download request
type ProfileRequest struct {
	EID              string `json:"eid"`
	ICCID            string `json:"iccid"`
	ProfileType      string `json:"profile_type"`
	ConfirmationCode string `json:"confirmation_code,omitempty"`
	PreferredCarrier string `json:"preferred_carrier,omitempty"` // Optional preferred carrier
	IMSI             string `json:"imsi,omitempty"`
	TenantID         string `json:"tenant_id,omitempty"`
}

// ProfileResponse represents the response from a profile operation
type ProfileResponse struct {
	Success        bool          `json:"success"`
	CarrierID      string        `json:"carrier_id"`
	ExecutionStatus string        `json:"execution_status"`
	StatusMessage   string        `json:"status_message"`
	ProfileState    string        `json:"profile_state,omitempty"`
	ResponseTime    time.Duration `json:"response_time"`
	RetriedOn       []string      `json:"retried_on,omitempty"` // List of carrier IDs tried before success
}

// ManagerConfig configures the SM-DP+ Manager
type ManagerConfig struct {
	HealthCheckInterval     time.Duration `json:"health_check_interval"`
	MaxRetries              int           `json:"max_retries"`
	RetryDelay              time.Duration `json:"retry_delay"`
	CircuitBreakerThreshold int          `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
	EnableLoadBalancing     bool          `json:"enable_load_balancing"`
	EnableFailover          bool          `json:"enable_failover"`
	DefaultTimeout          time.Duration `json:"default_timeout"`
}

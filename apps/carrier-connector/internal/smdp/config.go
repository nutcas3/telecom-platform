package smdp

import (
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
)

// DefaultManagerConfig returns default configuration for SM-DP+ Manager
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		HealthCheckInterval:     30 * time.Second,
		MaxRetries:              3,
		RetryDelay:              2 * time.Second,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   60 * time.Second,
		EnableLoadBalancing:     true,
		EnableFailover:          true,
		DefaultTimeout:          30 * time.Second,
	}
}

// CarrierConfig represents carrier configuration from database or config file
type CarrierConfig struct {
	ID                    string   `json:"id" yaml:"id"`
	Name                  string   `json:"name" yaml:"name"`
	CountryCode           string   `json:"country_code" yaml:"country_code"`
	MCC                   string   `json:"mcc" yaml:"mcc"`
	MNC                   string   `json:"mnc" yaml:"mnc"`
	ES2BaseURL            string   `json:"es2_base_url" yaml:"es2_base_url"`
	ES2APIKey             string   `json:"es2_api_key" yaml:"es2_api_key"`
	ES2InsecureSkip       bool     `json:"es2_insecure_skip" yaml:"es2_insecure_skip"`
	Priority              int      `json:"priority" yaml:"priority"`
	IsActive              bool     `json:"is_active" yaml:"is_active"`
	MaxConcurrentReqs     int      `json:"max_concurrent_requests" yaml:"max_concurrent_requests"`
	SupportedProfileTypes []string `json:"supported_profile_types" yaml:"supported_profile_types"`
	SupportedMCCs         []string `json:"supported_mccs" yaml:"supported_mccs"`
	SupportedMNCs         []string `json:"supported_mncs" yaml:"supported_mncs"`
	Features              []string `json:"features" yaml:"features"`
}

// ToCarrier converts CarrierConfig to Carrier struct
func (c *CarrierConfig) ToCarrier() *Carrier {
	carrier := &Carrier{
		ID:              c.ID,
		Name:            c.Name,
		CountryCode:     c.CountryCode,
		MCC:             c.MCC,
		MNC:             c.MNC,
		Priority:        c.Priority,
		IsActive:        c.IsActive,
		HealthStatus:    CarrierStatusUnknown,
		LastHealthCheck: time.Now(),
		Metrics: &CarrierMetrics{
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: 0,
			RequestRate:         0,
		},
		Capabilities: &CarrierCapabilities{
			SupportedProfileTypes: c.SupportedProfileTypes,
			MaxConcurrentRequests: c.MaxConcurrentReqs,
			SupportedMCCs:         c.SupportedMCCs,
			SupportedMNCs:         c.SupportedMNCs,
			Features:              c.Features,
		},
		ES2Config: &config.ES2Config{
			BaseURL:                  c.ES2BaseURL,
			APIKey:                   c.ES2APIKey,
			InsecureSkipVerify:       c.ES2InsecureSkip,
			FunctionalityRequesterID: "telecom-platform",
		},
	}

	return carrier
}

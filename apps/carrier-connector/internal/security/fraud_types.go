package security

import "time"

// FraudType represents different types of fraud
type FraudType string

const (
	FraudTypeAccountTakeover   FraudType = "account_takeover"
	FraudTypeSubscriptionFraud FraudType = "subscription_fraud"
	FraudTypePaymentFraud      FraudType = "payment_fraud"
	FraudTypeUsageAnomaly      FraudType = "usage_anomaly"
	FraudTypeSIMSwap           FraudType = "sim_swap"
)

// FraudSeverity represents the severity of fraud detection
type FraudSeverity string

const (
	FraudSeverityLow      FraudSeverity = "low"
	FraudSeverityMedium   FraudSeverity = "medium"
	FraudSeverityHigh     FraudSeverity = "high"
	FraudSeverityCritical FraudSeverity = "critical"
)

// FraudAlert represents a fraud detection alert
type FraudAlert struct {
	ID          string         `json:"id"`
	Type        FraudType      `json:"type"`
	Severity    FraudSeverity  `json:"severity"`
	ProfileID   string         `json:"profile_id"`
	Description string         `json:"description"`
	RiskScore   float64        `json:"risk_score"`
	Evidence    []string       `json:"evidence"`
	IPAddress   string         `json:"ip_address"`
	Timestamp   time.Time      `json:"timestamp"`
	Status      string         `json:"status"`
	Actions     []string       `json:"actions_taken"`
	Metadata    map[string]any `json:"metadata"`
}

// FraudPattern represents a fraud detection pattern
type FraudPattern struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      FraudType `json:"type"`
	Threshold float64   `json:"threshold"`
	Weight    float64   `json:"weight"`
	Enabled   bool      `json:"enabled"`
}

// FraudMetrics represents fraud detection metrics
type FraudMetrics struct {
	Period            string                  `json:"period"`
	TotalAlerts       int64                   `json:"total_alerts"`
	ResolvedAlerts    int64                   `json:"resolved_alerts"`
	FalsePositives    int64                   `json:"false_positives"`
	ResolutionRate    float64                 `json:"resolution_rate_pct"`
	FalsePositiveRate float64                 `json:"false_positive_rate_pct"`
	ByType            map[FraudType]int64     `json:"by_type"`
	BySeverity        map[FraudSeverity]int64 `json:"by_severity"`
	GeneratedAt       time.Time               `json:"generated_at"`
}

// FraudAlertFilter filters fraud alerts
type FraudAlertFilter struct {
	Type     FraudType     `json:"type,omitempty"`
	Severity FraudSeverity `json:"severity,omitempty"`
	Status   string        `json:"status,omitempty"`
	FromDate *time.Time    `json:"from_date,omitempty"`
	ToDate   *time.Time    `json:"to_date,omitempty"`
}

// FraudConfig configures the fraud detection service
type FraudConfig struct {
	EnableMLModels     bool
	AlertRetentionDays int
	AutoBlockThreshold float64
}

// DefaultFraudPatterns returns standard fraud detection patterns
func DefaultFraudPatterns() []FraudPattern {
	return []FraudPattern{
		{ID: "multiple_subs", Name: "Multiple Subscriptions", Type: FraudTypeSubscriptionFraud, Threshold: 3, Weight: 0.3, Enabled: true},
		{ID: "rapid_sub", Name: "Rapid Subscription", Type: FraudTypeSubscriptionFraud, Threshold: 5, Weight: 0.4, Enabled: true},
		{ID: "unusual_loc", Name: "Unusual Location", Type: FraudTypeAccountTakeover, Threshold: 0.8, Weight: 0.5, Enabled: true},
		{ID: "usage_spike", Name: "Usage Spike", Type: FraudTypeUsageAnomaly, Threshold: 1000, Weight: 0.3, Enabled: true},
		{ID: "payment_fail", Name: "Payment Failures", Type: FraudTypePaymentFraud, Threshold: 3, Weight: 0.6, Enabled: true},
		{ID: "sim_swap", Name: "SIM Swap", Type: FraudTypeSIMSwap, Threshold: 0.7, Weight: 0.8, Enabled: true},
	}
}

// SeverityFromScore determines severity from risk score
func SeverityFromScore(score float64) FraudSeverity {
	switch {
	case score >= 80:
		return FraudSeverityCritical
	case score >= 60:
		return FraudSeverityHigh
	case score >= 40:
		return FraudSeverityMedium
	default:
		return FraudSeverityLow
	}
}

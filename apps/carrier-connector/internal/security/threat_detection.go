package security

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ThreatLevel represents severity of detected threats
type ThreatLevel string

const (
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// ThreatType represents types of security threats
type ThreatType string

const (
	ThreatTypeBruteForce     ThreatType = "brute_force"
	ThreatTypeRateLimitAbuse ThreatType = "rate_limit_abuse"
	ThreatTypeSQLInjection   ThreatType = "sql_injection"
	ThreatTypeXSS            ThreatType = "xss"
	ThreatTypeUnauthorized   ThreatType = "unauthorized_access"
	ThreatTypeDataExfil      ThreatType = "data_exfiltration"
	ThreatTypeAnomalous      ThreatType = "anomalous_behavior"
)

// ThreatEvent represents a detected security threat
type ThreatEvent struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	Type        ThreatType     `json:"type"`
	Level       ThreatLevel    `json:"level"`
	Source      string         `json:"source"`
	Target      string         `json:"target"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
	DetectedAt  time.Time      `json:"detected_at"`
	Mitigated   bool           `json:"mitigated"`
	MitigatedAt *time.Time     `json:"mitigated_at"`
}

// ThreatDetector provides threat detection capabilities
type ThreatDetector struct {
	logger       *logrus.Logger
	rules        []DetectionRule
	events       []*ThreatEvent
	ipTracker    map[string]*IPActivity
	mu           sync.RWMutex
	alertHandler func(*ThreatEvent)
}

// IPActivity tracks activity per IP
type IPActivity struct {
	IP              string
	FailedLogins    int
	RequestCount    int
	LastRequest     time.Time
	FirstSeen       time.Time
	SuspiciousCount int
}

// DetectionRule defines a threat detection rule
type DetectionRule struct {
	ID          string
	Name        string
	Type        ThreatType
	Threshold   int
	Window      time.Duration
	Level       ThreatLevel
	Enabled     bool
}

// ThreatDetectorConfig configures the threat detector
type ThreatDetectorConfig struct {
	MaxFailedLogins    int
	RateLimitThreshold int
	AlertHandler       func(*ThreatEvent)
}

// DefaultThreatDetectorConfig returns default configuration
func DefaultThreatDetectorConfig() ThreatDetectorConfig {
	return ThreatDetectorConfig{
		MaxFailedLogins:    5,
		RateLimitThreshold: 1000,
	}
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector(logger *logrus.Logger, config ThreatDetectorConfig) *ThreatDetector {
	td := &ThreatDetector{
		logger:       logger,
		ipTracker:    make(map[string]*IPActivity),
		alertHandler: config.AlertHandler,
		rules:        defaultRules(config),
	}
	go td.cleanupLoop()
	return td
}

func defaultRules(config ThreatDetectorConfig) []DetectionRule {
	return []DetectionRule{
		{
			ID:        "brute_force_login",
			Name:      "Brute Force Login Detection",
			Type:      ThreatTypeBruteForce,
			Threshold: config.MaxFailedLogins,
			Window:    5 * time.Minute,
			Level:     ThreatLevelHigh,
			Enabled:   true,
		},
		{
			ID:        "rate_limit_abuse",
			Name:      "Rate Limit Abuse Detection",
			Type:      ThreatTypeRateLimitAbuse,
			Threshold: config.RateLimitThreshold,
			Window:    time.Minute,
			Level:     ThreatLevelMedium,
			Enabled:   true,
		},
	}
}

// RecordRequest records an API request for analysis
func (td *ThreatDetector) RecordRequest(ctx context.Context, ip, path, method string) {
	td.mu.Lock()
	defer td.mu.Unlock()

	activity, exists := td.ipTracker[ip]
	if !exists {
		activity = &IPActivity{
			IP:        ip,
			FirstSeen: time.Now(),
		}
		td.ipTracker[ip] = activity
	}

	activity.RequestCount++
	activity.LastRequest = time.Now()

	// Check rate limit rule
	for _, rule := range td.rules {
		if rule.Type == ThreatTypeRateLimitAbuse && rule.Enabled {
			if activity.RequestCount > rule.Threshold {
				td.raiseAlert(&ThreatEvent{
					Type:        ThreatTypeRateLimitAbuse,
					Level:       rule.Level,
					Source:      ip,
					Target:      path,
					Description: "Rate limit threshold exceeded",
					DetectedAt:  time.Now(),
				})
			}
		}
	}
}

// RecordFailedLogin records a failed login attempt
func (td *ThreatDetector) RecordFailedLogin(ctx context.Context, ip, userID string) {
	td.mu.Lock()
	defer td.mu.Unlock()

	activity, exists := td.ipTracker[ip]
	if !exists {
		activity = &IPActivity{
			IP:        ip,
			FirstSeen: time.Now(),
		}
		td.ipTracker[ip] = activity
	}

	activity.FailedLogins++
	activity.LastRequest = time.Now()

	// Check brute force rule
	for _, rule := range td.rules {
		if rule.Type == ThreatTypeBruteForce && rule.Enabled {
			if activity.FailedLogins >= rule.Threshold {
				td.raiseAlert(&ThreatEvent{
					Type:        ThreatTypeBruteForce,
					Level:       rule.Level,
					Source:      ip,
					Target:      userID,
					Description: "Multiple failed login attempts detected",
					DetectedAt:  time.Now(),
					Metadata:    map[string]any{"attempts": activity.FailedLogins},
				})
			}
		}
	}
}

// DetectSQLInjection checks for SQL injection patterns
func (td *ThreatDetector) DetectSQLInjection(_ context.Context, ip, input string) bool {
	patterns := []string{
		"'--", "'; DROP", "1=1", "OR 1=1", "UNION SELECT",
		"'; DELETE", "'; UPDATE", "'; INSERT",
	}

	for _, pattern := range patterns {
		if containsIgnoreCase(input, pattern) {
			td.raiseAlert(&ThreatEvent{
				Type:        ThreatTypeSQLInjection,
				Level:       ThreatLevelCritical,
				Source:      ip,
				Description: "SQL injection attempt detected",
				DetectedAt:  time.Now(),
				Metadata:    map[string]any{"pattern": pattern},
			})
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) // Simplified check
}

func (td *ThreatDetector) raiseAlert(event *ThreatEvent) {
	td.events = append(td.events, event)

	td.logger.WithFields(logrus.Fields{
		"type":   event.Type,
		"level":  event.Level,
		"source": event.Source,
	}).Warn("Security threat detected")

	if td.alertHandler != nil {
		td.alertHandler(event)
	}
}

// GetRecentEvents returns recent threat events
func (td *ThreatDetector) GetRecentEvents(limit int) []*ThreatEvent {
	td.mu.RLock()
	defer td.mu.RUnlock()

	if len(td.events) <= limit {
		return td.events
	}
	return td.events[len(td.events)-limit:]
}

// GetIPActivity returns activity for an IP
func (td *ThreatDetector) GetIPActivity(ip string) *IPActivity {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.ipTracker[ip]
}

// BlockIP marks an IP as blocked
func (td *ThreatDetector) BlockIP(ip string) {
	td.mu.Lock()
	defer td.mu.Unlock()

	if activity, exists := td.ipTracker[ip]; exists {
		activity.SuspiciousCount = 999
	}
}

// IsBlocked checks if an IP is blocked
func (td *ThreatDetector) IsBlocked(ip string) bool {
	td.mu.RLock()
	defer td.mu.RUnlock()

	if activity, exists := td.ipTracker[ip]; exists {
		return activity.SuspiciousCount >= 999
	}
	return false
}

func (td *ThreatDetector) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		td.cleanup()
	}
}

func (td *ThreatDetector) cleanup() {
	td.mu.Lock()
	defer td.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for ip, activity := range td.ipTracker {
		if activity.LastRequest.Before(cutoff) && activity.SuspiciousCount < 999 {
			delete(td.ipTracker, ip)
		}
	}
}

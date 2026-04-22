package monitoring

import (
	"context"
	"fmt"
	"log"
	"maps"
	"sync"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/websocket"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	StatusFiring   AlertStatus = "firing"
	StatusResolved AlertStatus = "resolved"
	StatusSilenced AlertStatus = "silenced"
)

// Alert represents a monitoring alert
type Alert struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Severity      AlertSeverity     `json:"severity"`
	Status        AlertStatus       `json:"status"`
	Timestamp     time.Time         `json:"timestamp"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	LastEvaluated time.Time         `json:"last_evaluated"`
}

// AlertRule defines a rule for generating alerts
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    AlertSeverity     `json:"severity"`
	Query       string            `json:"query"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Enabled     bool              `json:"enabled"`
}

// AlertManager manages alerts and alert rules
type AlertManager struct {
	rules     map[string]*AlertRule
	alerts    map[string]*Alert
	mutex     sync.RWMutex
	evaluator *AlertEvaluator
	notifier  AlertNotifier
}

// AlertEvaluator evaluates alert rules
type AlertEvaluator struct {
	healthMonitor *HealthMonitor
}

// AlertNotifier sends alert notifications
type AlertNotifier interface {
	SendAlert(alert *Alert) error
	ResolveAlert(alertID string) error
}

// WebSocketNotifier sends alerts via WebSocket
type WebSocketNotifier struct{}

func (wsn *WebSocketNotifier) SendAlert(alert *Alert) error {
	// Broadcast alert via WebSocket
	websocket.BroadcastAlertUpdate(alert.ID, string(alert.Severity), alert.Description, string(alert.Status))
	return nil
}

func (wsn *WebSocketNotifier) ResolveAlert(alertID string) error {
	// Broadcast alert resolution via WebSocket
	websocket.BroadcastAlertUpdate(alertID, "info", "Alert resolved", "resolved")
	return nil
}

// LogNotifier sends alerts to logs
type LogNotifier struct{}

func (ln *LogNotifier) SendAlert(alert *Alert) error {
	log.Printf("ALERT [%s] %s: %s", alert.Severity, alert.Name, alert.Description)
	return nil
}

func (ln *LogNotifier) ResolveAlert(alertID string) error {
	log.Printf("ALERT RESOLVED: %s", alertID)
	return nil
}

// CompositeNotifier sends alerts via multiple channels
type CompositeNotifier struct {
	notifiers []AlertNotifier
}

func (cn *CompositeNotifier) SendAlert(alert *Alert) error {
	for _, notifier := range cn.notifiers {
		if err := notifier.SendAlert(alert); err != nil {
			log.Printf("Failed to send alert via notifier: %v", err)
		}
	}
	return nil
}

func (cn *CompositeNotifier) ResolveAlert(alertID string) error {
	for _, notifier := range cn.notifiers {
		if err := notifier.ResolveAlert(alertID); err != nil {
			log.Printf("Failed to resolve alert via notifier: %v", err)
		}
	}
	return nil
}

// NewAlertManager creates a new alert manager
func NewAlertManager(healthMonitor *HealthMonitor) *AlertManager {
	evaluator := &AlertEvaluator{healthMonitor: healthMonitor}

	// Create composite notifier with WebSocket and log notifiers
	notifier := &CompositeNotifier{
		notifiers: []AlertNotifier{
			&WebSocketNotifier{},
			&LogNotifier{},
		},
	}

	return &AlertManager{
		rules:     make(map[string]*AlertRule),
		alerts:    make(map[string]*Alert),
		evaluator: evaluator,
		notifier:  notifier,
	}
}

// AddRule adds an alert rule
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.rules[rule.ID] = rule
}

// RemoveRule removes an alert rule
func (am *AlertManager) RemoveRule(ruleID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	delete(am.rules, ruleID)

	// Also remove any active alerts for this rule
	delete(am.alerts, ruleID)
}

// GetRules returns all alert rules
func (am *AlertManager) GetRules() map[string]*AlertRule {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	rules := make(map[string]*AlertRule)
	maps.Copy(rules, am.rules)
	return rules
}

// GetAlerts returns all active alerts
func (am *AlertManager) GetAlerts() map[string]*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	alerts := make(map[string]*Alert)
	maps.Copy(alerts, am.alerts)
	return alerts
}

// EvaluateRules evaluates all alert rules
func (am *AlertManager) EvaluateRules(ctx context.Context) {
	am.mutex.RLock()
	rules := make(map[string]*AlertRule)
	maps.Copy(rules, am.rules)
	am.mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		shouldAlert, value, err := am.evaluator.EvaluateRule(ctx, rule)
		if err != nil {
			log.Printf("Error evaluating rule %s: %v", rule.ID, err)
			continue
		}

		am.handleRuleEvaluation(rule, shouldAlert, value)
	}
}

// handleRuleEvaluation handles the result of a rule evaluation
func (am *AlertManager) handleRuleEvaluation(rule *AlertRule, shouldAlert bool, value float64) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	alert, exists := am.alerts[rule.ID]

	if shouldAlert && (!exists || alert.Status == StatusResolved) {
		// Create or update alert
		newAlert := &Alert{
			ID:            rule.ID,
			Name:          rule.Name,
			Description:   fmt.Sprintf("%s (current: %.2f, threshold: %.2f)", rule.Description, value, rule.Threshold),
			Severity:      rule.Severity,
			Status:        StatusFiring,
			Timestamp:     time.Now(),
			LastEvaluated: time.Now(),
			Labels:        rule.Labels,
			Annotations: map[string]string{
				"threshold": fmt.Sprintf("%.2f", rule.Threshold),
				"value":     fmt.Sprintf("%.2f", value),
				"query":     rule.Query,
			},
		}

		am.alerts[rule.ID] = newAlert

		// Send notification
		if err := am.notifier.SendAlert(newAlert); err != nil {
			log.Printf("Failed to send alert notification: %v", err)
		}

	} else if !shouldAlert && exists && alert.Status == StatusFiring {
		// Resolve alert
		alert.Status = StatusResolved
		alert.LastEvaluated = time.Now()

		// Send resolution notification
		if err := am.notifier.ResolveAlert(alert.ID); err != nil {
			log.Printf("Failed to send alert resolution: %v", err)
		}
	}
}

// Start starts the alert manager evaluation loop
func (am *AlertManager) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.EvaluateRules(ctx)
		}
	}
}

// EvaluateRule evaluates a single alert rule
func (ae *AlertEvaluator) EvaluateRule(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	switch rule.Query {
	case "health_check":
		return ae.evaluateHealthCheck(ctx, rule)
	case "system_memory":
		return ae.evaluateSystemMemory(ctx, rule)
	case "system_goroutines":
		return ae.evaluateSystemGoroutines(ctx, rule)
	case "database_connections":
		return ae.evaluateDatabaseConnections(ctx, rule)
	default:
		return false, 0, fmt.Errorf("unknown alert query: %s", rule.Query)
	}
}

// evaluateHealthCheck evaluates health check based alerts
func (ae *AlertEvaluator) evaluateHealthCheck(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	if ae.healthMonitor == nil {
		return false, 0, fmt.Errorf("health monitor not available")
	}

	health := ae.healthMonitor.CheckHealth(ctx)

	// Convert health status to numeric value for threshold comparison
	var value float64
	switch health.Status {
	case StatusHealthy:
		value = 0
	case StatusDegraded:
		value = 1
	case StatusUnhealthy:
		value = 2
	}

	// Alert if health status meets or exceeds threshold
	return value >= rule.Threshold, value, nil
}

// evaluateSystemMemory evaluates system memory usage
func (ae *AlertEvaluator) evaluateSystemMemory(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	systemChecker := NewSystemHealthChecker()
	check := systemChecker.CheckHealth(ctx)

	// Extract memory usage percentage from metadata
	memoryUsagePercent := 0.0
	if metadata, ok := check.Metadata["memory_usage_percent"].(float64); ok {
		memoryUsagePercent = metadata
	}

	return memoryUsagePercent >= rule.Threshold, memoryUsagePercent, nil
}

// evaluateSystemGoroutines evaluates goroutine count
func (ae *AlertEvaluator) evaluateSystemGoroutines(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	systemChecker := NewSystemHealthChecker()
	check := systemChecker.CheckHealth(ctx)

	// Extract goroutine count from metadata
	goroutines := 0.0
	if metadata, ok := check.Metadata["goroutines"].(int); ok {
		goroutines = float64(metadata)
	}

	return goroutines >= rule.Threshold, goroutines, nil
}

// evaluateDatabaseConnections evaluates database connection pool
func (ae *AlertEvaluator) evaluateDatabaseConnections(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	if ae.healthMonitor == nil {
		return false, 0, fmt.Errorf("health monitor not available")
	}

	// Get database health check result which contains connection pool metadata
	health := ae.healthMonitor.CheckHealth(ctx)
	dbCheck, exists := health.Checks["database"]
	if !exists {
		return false, 0, fmt.Errorf("database health check not registered")
	}

	// Calculate connection pool utilization percentage
	openConns, ok1 := dbCheck.Metadata["open_connections"].(int)
	inUse, ok2 := dbCheck.Metadata["in_use"].(int)
	if !ok1 || !ok2 || openConns == 0 {
		return false, 0, nil
	}

	utilization := float64(inUse) / float64(openConns) * 100
	return utilization >= rule.Threshold, utilization, nil
}

// Global alert manager instance
var globalAlertManager *AlertManager

// InitializeAlertManager initializes the global alert manager
func InitializeAlertManager(healthMonitor *HealthMonitor) {
	globalAlertManager = NewAlertManager(healthMonitor)

	// Add default alert rules
	globalAlertManager.AddRule(&AlertRule{
		ID:          "system_health",
		Name:        "System Health Check",
		Description: "System health is degraded or unhealthy",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   1, // Alert if status is degraded (1) or unhealthy (2)
		Duration:    time.Minute * 5,
		Labels: map[string]string{
			"component": "system",
			"check":     "health",
		},
		Enabled: true,
	})

	globalAlertManager.AddRule(&AlertRule{
		ID:          "high_memory_usage",
		Name:        "High Memory Usage",
		Description: "System memory usage is high",
		Severity:    SeverityWarning,
		Query:       "system_memory",
		Threshold:   80, // Alert if memory usage > 80%
		Duration:    time.Minute * 5,
		Labels: map[string]string{
			"component": "system",
			"metric":    "memory",
		},
		Enabled: true,
	})

	globalAlertManager.AddRule(&AlertRule{
		ID:          "high_goroutine_count",
		Name:        "High Goroutine Count",
		Description: "System has too many goroutines",
		Severity:    SeverityWarning,
		Query:       "system_goroutines",
		Threshold:   1000, // Alert if goroutines > 1000
		Duration:    time.Minute * 5,
		Labels: map[string]string{
			"component": "system",
			"metric":    "goroutines",
		},
		Enabled: true,
	})
}

// GetAlertManager returns the global alert manager
func GetAlertManager() *AlertManager {
	return globalAlertManager
}

// StartAlertManager starts the global alert manager
func StartAlertManager(ctx context.Context) {
	if globalAlertManager != nil {
		go globalAlertManager.Start(ctx, time.Minute*1) // Evaluate every minute
	}
}

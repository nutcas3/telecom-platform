package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlertNotifier is a mock implementation of AlertNotifier
type MockAlertNotifier struct {
	mock.Mock
}

func (m *MockAlertNotifier) SendAlert(alert *Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

func (m *MockAlertNotifier) ResolveAlert(alertID string) error {
	args := m.Called(alertID)
	return args.Error(0)
}

func TestAlertManager_AddRule(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	rule := &AlertRule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "Test description",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   1,
		Duration:    time.Minute,
		Enabled:     true,
	}

	am.AddRule(rule)

	rules := am.GetRules()
	assert.Len(t, rules, 1)
	assert.Equal(t, rule, rules["test-rule"])
}

func TestAlertManager_RemoveRule(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	rule := &AlertRule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "Test description",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   1,
		Duration:    time.Minute,
		Enabled:     true,
	}

	am.AddRule(rule)
	assert.Len(t, am.GetRules(), 1)

	am.RemoveRule("test-rule")
	assert.Len(t, am.GetRules(), 0)
}

func TestAlertManager_EvaluateRules_HealthCheck(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	rule := &AlertRule{
		ID:          "health-rule",
		Name:        "Health Check",
		Description: "System health check",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   1, // Alert if degraded (1) or unhealthy (2)
		Duration:    time.Minute,
		Enabled:     true,
	}

	am.AddRule(rule)

	ctx := context.Background()
	am.EvaluateRules(ctx)

	// Should have an alert since system health is healthy (0 < 1 threshold)
	alerts := am.GetAlerts()
	assert.Len(t, alerts, 0) // No alert since healthy (0) < threshold (1)
}

func TestAlertManager_EvaluateRules_SystemMemory(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	rule := &AlertRule{
		ID:          "memory-rule",
		Name:        "Memory Usage",
		Description: "High memory usage",
		Severity:    SeverityWarning,
		Query:       "system_memory",
		Threshold:   99, // Very high threshold to avoid false positives
		Duration:    time.Minute,
		Enabled:     true,
	}

	am.AddRule(rule)

	ctx := context.Background()
	am.EvaluateRules(ctx)

	// Should not have an alert since memory usage is typically < 99%
	alerts := am.GetAlerts()
	assert.Len(t, alerts, 0)
}

func TestAlertManager_EvaluateRules_DisabledRule(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	rule := &AlertRule{
		ID:          "disabled-rule",
		Name:        "Disabled Rule",
		Description: "This rule is disabled",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   0, // Would always alert
		Duration:    time.Minute,
		Enabled:     false, // Disabled
	}

	am.AddRule(rule)

	ctx := context.Background()
	am.EvaluateRules(ctx)

	// Should not have an alert since rule is disabled
	alerts := am.GetAlerts()
	assert.Len(t, alerts, 0)
}

func TestAlertEvaluator_EvaluateRule_HealthCheck(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	evaluator := &AlertEvaluator{healthMonitor: hm}

	rule := &AlertRule{
		ID:        "health-rule",
		Query:     "health_check",
		Threshold: 1,
	}

	ctx := context.Background()
	shouldAlert, value, err := evaluator.EvaluateRule(ctx, rule)

	assert.NoError(t, err)
	assert.False(t, shouldAlert) // Healthy (0) < threshold (1)
	assert.Equal(t, 0.0, value)
}

func TestAlertEvaluator_EvaluateRule_SystemMemory(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	evaluator := &AlertEvaluator{healthMonitor: hm}

	rule := &AlertRule{
		ID:        "memory-rule",
		Query:     "system_memory",
		Threshold: 50,
	}

	ctx := context.Background()
	shouldAlert, value, err := evaluator.EvaluateRule(ctx, rule)

	assert.NoError(t, err)
	// Memory usage is typically low, so should not alert
	assert.False(t, shouldAlert)
	assert.Greater(t, value, 0.0)
}

func TestAlertEvaluator_EvaluateRule_UnknownQuery(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	evaluator := &AlertEvaluator{healthMonitor: hm}

	rule := &AlertRule{
		ID:        "unknown-rule",
		Query:     "unknown_query",
		Threshold: 1,
	}

	ctx := context.Background()
	shouldAlert, value, err := evaluator.EvaluateRule(ctx, rule)

	assert.Error(t, err)
	assert.False(t, shouldAlert)
	assert.Equal(t, 0.0, value)
	assert.Contains(t, err.Error(), "unknown alert query")
}

func TestWebSocketNotifier_SendAlert(t *testing.T) {
	notifier := &WebSocketNotifier{}

	alert := &Alert{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test description",
		Severity:    SeverityWarning,
		Status:      StatusFiring,
		Timestamp:   time.Now(),
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
}

func TestWebSocketNotifier_ResolveAlert(t *testing.T) {
	notifier := &WebSocketNotifier{}

	err := notifier.ResolveAlert("test-alert")
	assert.NoError(t, err)
}

func TestLogNotifier_SendAlert(t *testing.T) {
	notifier := &LogNotifier{}

	alert := &Alert{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test description",
		Severity:    SeverityWarning,
		Status:      StatusFiring,
		Timestamp:   time.Now(),
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
}

func TestLogNotifier_ResolveAlert(t *testing.T) {
	notifier := &LogNotifier{}

	err := notifier.ResolveAlert("test-alert")
	assert.NoError(t, err)
}

func TestCompositeNotifier_SendAlert(t *testing.T) {
	mockNotifier1 := &MockAlertNotifier{}
	mockNotifier2 := &MockAlertNotifier{}

	mockNotifier1.On("SendAlert", mock.AnythingOfType("*monitoring.Alert")).Return(nil)
	mockNotifier2.On("SendAlert", mock.AnythingOfType("*monitoring.Alert")).Return(nil)

	notifier := &CompositeNotifier{
		notifiers: []AlertNotifier{mockNotifier1, mockNotifier2},
	}

	alert := &Alert{
		ID:          "test-alert",
		Name:        "Test Alert",
		Description: "Test description",
		Severity:    SeverityWarning,
		Status:      StatusFiring,
		Timestamp:   time.Now(),
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)

	mockNotifier1.AssertExpectations(t)
	mockNotifier2.AssertExpectations(t)
}

func TestCompositeNotifier_ResolveAlert(t *testing.T) {
	mockNotifier1 := &MockAlertNotifier{}
	mockNotifier2 := &MockAlertNotifier{}

	mockNotifier1.On("ResolveAlert", "test-alert").Return(nil)
	mockNotifier2.On("ResolveAlert", "test-alert").Return(nil)

	notifier := &CompositeNotifier{
		notifiers: []AlertNotifier{mockNotifier1, mockNotifier2},
	}

	err := notifier.ResolveAlert("test-alert")
	assert.NoError(t, err)

	mockNotifier1.AssertExpectations(t)
	mockNotifier2.AssertExpectations(t)
}

func TestAlertRule_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rule    *AlertRule
		isValid bool
	}{
		{
			name: "Valid rule",
			rule: &AlertRule{
				ID:          "valid-rule",
				Name:        "Valid Rule",
				Description: "Valid description",
				Severity:    SeverityWarning,
				Query:       "health_check",
				Threshold:   1,
				Duration:    time.Minute,
				Enabled:     true,
			},
			isValid: true,
		},
		{
			name: "Invalid severity",
			rule: &AlertRule{
				ID:          "invalid-severity",
				Name:        "Invalid Severity",
				Description: "Invalid severity",
				Severity:    "invalid", // Invalid severity
				Query:       "health_check",
				Threshold:   1,
				Duration:    time.Minute,
				Enabled:     true,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check if required fields are set and severity is valid
			validSeverities := map[string]bool{
				string(SeverityInfo):     true,
				string(SeverityWarning):  true,
				string(SeverityCritical): true,
			}
			isValid := tt.rule.ID != "" && tt.rule.Name != "" && tt.rule.Query != "" && validSeverities[string(tt.rule.Severity)]
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestAlertStatus_Values(t *testing.T) {
	assert.Equal(t, AlertStatus("firing"), StatusFiring)
	assert.Equal(t, AlertStatus("resolved"), StatusResolved)
	assert.Equal(t, AlertStatus("silenced"), StatusSilenced)
}

func TestAlertSeverity_Values(t *testing.T) {
	assert.Equal(t, AlertSeverity("info"), SeverityInfo)
	assert.Equal(t, AlertSeverity("warning"), SeverityWarning)
	assert.Equal(t, AlertSeverity("critical"), SeverityCritical)
}

func TestAlertManager_Start(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	// Add a rule
	rule := &AlertRule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "Test description",
		Severity:    SeverityWarning,
		Query:       "health_check",
		Threshold:   1,
		Duration:    time.Minute,
		Enabled:     true,
	}
	am.AddRule(rule)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the alert manager (should run for 2 seconds then stop)
	am.Start(ctx, time.Second*1)

	// Should have completed without panicking
	select {
	case <-ctx.Done():
		// Expected - context timed out
	default:
		// Should have finished
	}
}

// Benchmark tests
func BenchmarkAlertManager_EvaluateRules(b *testing.B) {
	hm := NewHealthMonitor("1.0.0", "test")
	am := NewAlertManager(hm)

	// Add multiple rules
	for i := range 10 {
		rule := &AlertRule{
			ID:          fmt.Sprintf("rule-%d", i),
			Name:        fmt.Sprintf("Rule %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Severity:    SeverityWarning,
			Query:       "health_check",
			Threshold:   1,
			Duration:    time.Minute,
			Enabled:     true,
		}
		am.AddRule(rule)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.EvaluateRules(ctx)
	}
}

func BenchmarkAlertEvaluator_EvaluateRule(b *testing.B) {
	hm := NewHealthMonitor("1.0.0", "test")
	evaluator := &AlertEvaluator{healthMonitor: hm}

	rule := &AlertRule{
		ID:        "benchmark-rule",
		Query:     "health_check",
		Threshold: 1,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.EvaluateRule(ctx, rule)
	}
}

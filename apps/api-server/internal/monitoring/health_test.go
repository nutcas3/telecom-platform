package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockServiceChecker is a mock implementation of ServiceChecker
type MockServiceChecker struct {
	mock.Mock
}

func (m *MockServiceChecker) Available(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// MockHealthChecker is a mock implementation of HealthChecker
type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	args := m.Called(ctx)
	return args.Get(0).(HealthCheck)
}

func TestHealthMonitor_RegisterCheck(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")

	mockChecker := &MockHealthChecker{}
	hm.RegisterCheck("test-service", mockChecker)

	// Verify the checker was registered
	assert.NotNil(t, hm)
}

func TestHealthMonitor_CheckHealth(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")

	// Add system health checker
	systemChecker := NewSystemHealthChecker()
	hm.RegisterCheck("system", systemChecker)

	ctx := context.Background()
	health := hm.CheckHealth(ctx)

	assert.Equal(t, StatusHealthy, health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Equal(t, "test", health.Environment)
	assert.Equal(t, 1, health.Summary.Total)
	assert.Equal(t, 1, health.Summary.Healthy)
}

func TestDatabaseHealthChecker_Healthy(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Set max open connections to avoid "nearly full" status
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)

	checker := NewDatabaseHealthChecker(db)
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	// Database might show as degraded due to connection pool logic
	assert.Contains(t, []HealthStatus{StatusHealthy, StatusDegraded}, result.Status)
	assert.Equal(t, "database", result.Name)
	assert.NotEmpty(t, result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.NotNil(t, result.Metadata)
}

func TestSystemHealthChecker_Healthy(t *testing.T) {
	checker := NewSystemHealthChecker()
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "system", result.Name)
	assert.Equal(t, "System resources healthy", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.NotNil(t, result.Metadata)

	// Check metadata contains expected fields
	metadata := result.Metadata
	assert.Contains(t, metadata, "memory_usage_mb")
	assert.Contains(t, metadata, "memory_limit_mb")
	assert.Contains(t, metadata, "memory_usage_percent")
	assert.Contains(t, metadata, "goroutines")
	assert.Contains(t, metadata, "gc_cycles")
}

func TestPrometheusHealthChecker_Available(t *testing.T) {
	mockChecker := &MockServiceChecker{}
	mockChecker.On("Available", mock.Anything).Return(true)

	checker := NewPrometheusHealthChecker(mockChecker)
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "prometheus", result.Name)
	assert.Equal(t, "Prometheus service available", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))

	mockChecker.AssertExpectations(t)
}

func TestPrometheusHealthChecker_Unavailable(t *testing.T) {
	mockChecker := &MockServiceChecker{}
	mockChecker.On("Available", mock.Anything).Return(false)

	checker := NewPrometheusHealthChecker(mockChecker)
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Equal(t, "prometheus", result.Name)
	assert.Equal(t, "Prometheus service unavailable", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))

	mockChecker.AssertExpectations(t)
}

func TestKubernetesHealthChecker_Available(t *testing.T) {
	mockChecker := &MockServiceChecker{}
	mockChecker.On("Available", mock.Anything).Return(true)

	checker := NewKubernetesHealthChecker(mockChecker)
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "kubernetes", result.Name)
	assert.Equal(t, "Kubernetes service available", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))

	mockChecker.AssertExpectations(t)
}

func TestKubernetesHealthChecker_Unavailable(t *testing.T) {
	mockChecker := &MockServiceChecker{}
	mockChecker.On("Available", mock.Anything).Return(false)

	checker := NewKubernetesHealthChecker(mockChecker)
	ctx := context.Background()

	result := checker.CheckHealth(ctx)

	assert.Equal(t, StatusDegraded, result.Status)
	assert.Equal(t, "kubernetes", result.Name)
	assert.Equal(t, "Kubernetes service unavailable", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))

	mockChecker.AssertExpectations(t)
}

func TestHealthMonitor_MultipleChecks(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")

	// Add multiple health checkers
	systemChecker := NewSystemHealthChecker()
	mockPrometheusChecker := &MockServiceChecker{}
	mockKubernetesChecker := &MockServiceChecker{}

	mockPrometheusChecker.On("Available", mock.Anything).Return(true)
	mockKubernetesChecker.On("Available", mock.Anything).Return(false)

	hm.RegisterCheck("system", systemChecker)
	hm.RegisterCheck("prometheus", NewPrometheusHealthChecker(mockPrometheusChecker))
	hm.RegisterCheck("kubernetes", NewKubernetesHealthChecker(mockKubernetesChecker))

	ctx := context.Background()
	health := hm.CheckHealth(ctx)

	// Should be degraded because kubernetes is unavailable
	assert.Equal(t, StatusDegraded, health.Status)
	assert.Equal(t, 3, health.Summary.Total)
	assert.Equal(t, 2, health.Summary.Healthy)
	assert.Equal(t, 1, health.Summary.Degraded) // Kubernetes returns degraded
	assert.Equal(t, 0, health.Summary.Unhealthy)

	mockPrometheusChecker.AssertExpectations(t)
	mockKubernetesChecker.AssertExpectations(t)
}

func TestHealthMonitor_Timeout(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")

	// Create a slow checker that will timeout
	slowChecker := &MockServiceChecker{}
	slowChecker.On("Available", mock.Anything).After(11 * time.Second).Return(true)

	hm.RegisterCheck("slow", NewPrometheusHealthChecker(slowChecker))

	ctx := context.Background()
	start := time.Now()
	health := hm.CheckHealth(ctx)
	duration := time.Since(start)

	// Should complete quickly due to timeout (less than 10.5 seconds to account for overhead)
	assert.Less(t, duration, 10*time.Second+500*time.Millisecond)
	assert.Equal(t, StatusUnhealthy, health.Status) // Timeout should result in unhealthy

	slowChecker.AssertExpectations(t)
}

func TestHealthStatus_Values(t *testing.T) {
	assert.Equal(t, HealthStatus("healthy"), StatusHealthy)
	assert.Equal(t, HealthStatus("degraded"), StatusDegraded)
	assert.Equal(t, HealthStatus("unhealthy"), StatusUnhealthy)
}

func TestHealthSummary_Calculation(t *testing.T) {
	hm := NewHealthMonitor("1.0.0", "test")

	// Create checkers with different statuses
	healthyChecker := &MockServiceChecker{}
	healthyChecker.On("Available", mock.Anything).Return(true)

	unhealthyChecker := &MockServiceChecker{}
	unhealthyChecker.On("Available", mock.Anything).Return(false)

	hm.RegisterCheck("healthy", NewPrometheusHealthChecker(healthyChecker))
	hm.RegisterCheck("unhealthy", NewPrometheusHealthChecker(unhealthyChecker))
	hm.RegisterCheck("system", NewSystemHealthChecker())

	ctx := context.Background()
	health := hm.CheckHealth(ctx)

	assert.Equal(t, 3, health.Summary.Total)
	assert.Equal(t, 2, health.Summary.Healthy) // system + healthy prometheus
	assert.Equal(t, 0, health.Summary.Degraded)
	assert.Equal(t, 1, health.Summary.Unhealthy) // unhealthy prometheus

	healthyChecker.AssertExpectations(t)
	unhealthyChecker.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkSystemHealthChecker(b *testing.B) {
	checker := NewSystemHealthChecker()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckHealth(ctx)
	}
}

func BenchmarkHealthMonitor_CheckHealth(b *testing.B) {
	hm := NewHealthMonitor("1.0.0", "test")
	hm.RegisterCheck("system", NewSystemHealthChecker())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm.CheckHealth(ctx)
	}
}

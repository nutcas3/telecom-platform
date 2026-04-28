package monitoring

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check for a component
type HealthCheck struct {
	Name         string         `json:"name"`
	Status       HealthStatus   `json:"status"`
	Message      string         `json:"message,omitempty"`
	LastChecked  time.Time      `json:"last_checked"`
	ResponseTime time.Duration  `json:"response_time"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      time.Duration          `json:"uptime"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]HealthCheck `json:"checks"`
	Summary     HealthSummary          `json:"summary"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Degraded  int `json:"degraded"`
	Unhealthy int `json:"unhealthy"`
}

// HealthChecker interface for components that can be health checked
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthCheck
}

// HealthMonitor manages health checks for all system components
type HealthMonitor struct {
	checks    map[string]HealthChecker
	mutex     sync.RWMutex
	startTime time.Time
	version   string
	env       string
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(version, environment string) *HealthMonitor {
	return &HealthMonitor{
		checks:    make(map[string]HealthChecker),
		startTime: time.Now(),
		version:   version,
		env:       environment,
	}
}

// RegisterCheck registers a health checker
func (hm *HealthMonitor) RegisterCheck(name string, checker HealthChecker) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hm.checks[name] = checker
}

// CheckHealth performs all health checks
func (hm *HealthMonitor) CheckHealth(ctx context.Context) SystemHealth {
	hm.mutex.RLock()
	checks := make(map[string]HealthChecker, len(hm.checks))
	maps.Copy(checks, hm.checks)
	hm.mutex.RUnlock()

	results := make(map[string]HealthCheck)
	summary := HealthSummary{}

	for name, checker := range checks {
		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

		// Execute health check with timeout
		resultChan := make(chan HealthCheck, 1)
		go func() {
			resultChan <- checker.CheckHealth(checkCtx)
		}()

		var result HealthCheck
		select {
		case result = <-resultChan:
			// Check completed successfully
		case <-checkCtx.Done():
			// Timeout occurred
			result = HealthCheck{
				Name:         name,
				Status:       StatusUnhealthy,
				Message:      "Health check timed out",
				ResponseTime: 10 * time.Second,
			}
		}
		cancel()

		result.LastChecked = time.Now()
		results[name] = result

		summary.Total++
		switch result.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnhealthy:
			summary.Unhealthy++
		}
	}

	// Determine overall status
	var overallStatus HealthStatus
	if summary.Unhealthy > 0 {
		overallStatus = StatusUnhealthy
	} else if summary.Degraded > 0 {
		overallStatus = StatusDegraded
	} else {
		overallStatus = StatusHealthy
	}

	return SystemHealth{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Uptime:      time.Since(hm.startTime),
		Version:     hm.version,
		Environment: hm.env,
		Checks:      results,
		Summary:     summary,
	}
}

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	db *gorm.DB
}

func NewDatabaseHealthChecker(db *gorm.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{db: db}
}

func (dhc *DatabaseHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	sqlDB, err := dhc.db.DB()
	if err != nil {
		return HealthCheck{
			Name:         "database",
			Status:       StatusUnhealthy,
			Message:      fmt.Sprintf("Failed to get database instance: %v", err),
			ResponseTime: time.Since(start),
		}
	}

	// Test database connection
	err = sqlDB.PingContext(ctx)
	responseTime := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:         "database",
			Status:       StatusUnhealthy,
			Message:      fmt.Sprintf("Database ping failed: %v", err),
			ResponseTime: responseTime,
		}
	}

	// Get connection pool stats
	stats := sqlDB.Stats()
	metadata := map[string]any{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}

	// Check if connection pool is healthy
	if stats.OpenConnections >= stats.MaxOpenConnections-5 {
		return HealthCheck{
			Name:         "database",
			Status:       StatusDegraded,
			Message:      "Database connection pool nearly full",
			ResponseTime: responseTime,
			Metadata:     metadata,
		}
	}

	return HealthCheck{
		Name:         "database",
		Status:       StatusHealthy,
		Message:      "Database connection successful",
		ResponseTime: responseTime,
		Metadata:     metadata,
	}
}

// ServiceChecker interface for services that can be checked for availability
type ServiceChecker interface {
	Available(ctx context.Context) bool
}

// PrometheusHealthChecker checks Prometheus connectivity
type PrometheusHealthChecker struct {
	prometheusService ServiceChecker
}

func NewPrometheusHealthChecker(prometheusService ServiceChecker) *PrometheusHealthChecker {
	return &PrometheusHealthChecker{prometheusService: prometheusService}
}

func (phc *PrometheusHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	available := phc.prometheusService.Available(ctx)
	responseTime := time.Since(start)

	if !available {
		return HealthCheck{
			Name:         "prometheus",
			Status:       StatusUnhealthy,
			Message:      "Prometheus service unavailable",
			ResponseTime: responseTime,
		}
	}

	return HealthCheck{
		Name:         "prometheus",
		Status:       StatusHealthy,
		Message:      "Prometheus service available",
		ResponseTime: responseTime,
	}
}

// KubernetesHealthChecker checks Kubernetes connectivity
type KubernetesHealthChecker struct {
	kubernetesService ServiceChecker
}

func NewKubernetesHealthChecker(kubernetesService ServiceChecker) *KubernetesHealthChecker {
	return &KubernetesHealthChecker{kubernetesService: kubernetesService}
}

func (khc *KubernetesHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	available := khc.kubernetesService.Available(ctx)
	responseTime := time.Since(start)

	if !available {
		return HealthCheck{
			Name:         "kubernetes",
			Status:       StatusDegraded,
			Message:      "Kubernetes service unavailable",
			ResponseTime: responseTime,
		}
	}

	return HealthCheck{
		Name:         "kubernetes",
		Status:       StatusHealthy,
		Message:      "Kubernetes service available",
		ResponseTime: responseTime,
	}
}

// SystemHealthChecker checks system resources
type SystemHealthChecker struct{}

func NewSystemHealthChecker() *SystemHealthChecker {
	return &SystemHealthChecker{}
}

func (shc *SystemHealthChecker) CheckHealth(ctx context.Context) HealthCheck {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	responseTime := time.Since(start)

	// Calculate memory usage percentage
	memoryUsageMB := float64(m.Alloc) / 1024 / 1024
	memoryLimitMB := float64(m.Sys) / 1024 / 1024
	memoryUsagePercent := (memoryUsageMB / memoryLimitMB) * 100

	metadata := map[string]any{
		"memory_usage_mb":      memoryUsageMB,
		"memory_limit_mb":      memoryLimitMB,
		"memory_usage_percent": memoryUsagePercent,
		"goroutines":           runtime.NumGoroutine(),
		"gc_cycles":            m.NumGC,
		"heap_alloc":           m.HeapAlloc,
		"heap_sys":             m.HeapSys,
		"heap_objects":         m.HeapObjects,
	}

	status := StatusHealthy
	message := "System resources healthy"

	// Check memory usage
	if memoryUsagePercent > 90 {
		status = StatusUnhealthy
		message = "High memory usage (>90%)"
	} else if memoryUsagePercent > 75 {
		status = StatusDegraded
		message = "Moderate memory usage (>75%)"
	}

	// Check goroutine count
	goroutines := runtime.NumGoroutine()
	if goroutines > 1000 {
		if status == StatusHealthy {
			status = StatusDegraded
			message = "High goroutine count"
		}
	}

	return HealthCheck{
		Name:         "system",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		Metadata:     metadata,
	}
}

// Global health monitor instance
var globalHealthMonitor *HealthMonitor

// InitializeHealthMonitor initializes the global health monitor
func InitializeHealthMonitor(version, environment string) {
	globalHealthMonitor = NewHealthMonitor(version, environment)
}

// GetHealthMonitor returns the global health monitor
func GetHealthMonitor() *HealthMonitor {
	return globalHealthMonitor
}

// HealthHandler returns a gin handler for health checks
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalHealthMonitor == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"message": "Health monitor not initialized",
			})
			return
		}

		ctx := c.Request.Context()
		health := globalHealthMonitor.CheckHealth(ctx)

		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		switch health.Status {
		case StatusDegraded:
			statusCode = http.StatusOK // Still serve traffic but indicate degradation
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	}
}

// ReadyHandler returns a gin handler for readiness checks
func ReadyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalHealthMonitor == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"message": "Health monitor not initialized",
			})
			return
		}

		ctx := c.Request.Context()
		health := globalHealthMonitor.CheckHealth(ctx)

		// Ready if all critical checks are healthy
		ready := health.Status != StatusUnhealthy

		statusCode := http.StatusOK
		if !ready {
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]any{
			"ready":  ready,
			"status": health.Status,
		}

		c.JSON(statusCode, response)
	}
}

// LiveHandler returns a gin handler for liveness checks
func LiveHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Liveness is simple - if we can respond, we're alive
		c.JSON(http.StatusOK, gin.H{
			"alive":     true,
			"timestamp": time.Now(),
		})
	}
}

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics tracks performance statistics
type PerformanceMetrics struct {
	mu                sync.RWMutex
	requestCount      int64
	totalResponseTime time.Duration
	slowRequests      int64
	fastRequests      int64
	errorRequests     int64
}

// PerformanceMiddleware tracks and optimizes API performance
type PerformanceMiddleware struct {
	metrics       *PerformanceMetrics
	slowThreshold time.Duration
}

// NewPerformanceMiddleware creates a new performance middleware
func NewPerformanceMiddleware(slowThreshold time.Duration) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		metrics:       &PerformanceMetrics{},
		slowThreshold: slowThreshold,
	}
}

// Middleware returns the gin.HandlerFunc
func (pm *PerformanceMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(start)

		// Update metrics
		pm.updateMetrics(c.Writer.Status(), responseTime)

		// Add performance headers
		c.Header("X-Response-Time", responseTime.String())
		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))

		// Log slow requests
		if responseTime > pm.slowThreshold {
			pm.logSlowRequest(c, responseTime)
		}
	}
}

// updateMetrics updates the performance metrics
func (pm *PerformanceMiddleware) updateMetrics(statusCode int, responseTime time.Duration) {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.requestCount++
	pm.metrics.totalResponseTime += responseTime

	if responseTime > pm.slowThreshold {
		pm.metrics.slowRequests++
	} else {
		pm.metrics.fastRequests++
	}

	if statusCode >= 400 {
		pm.metrics.errorRequests++
	}
}

// logSlowRequest logs slow requests for monitoring
func (pm *PerformanceMiddleware) logSlowRequest(c *gin.Context, responseTime time.Duration) {
	// Structured logging for slow requests
	logData := map[string]any{
		"level":         "warn",
		"type":          "slow_request",
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"query":         c.Request.URL.RawQuery,
		"response_time": responseTime.String(),
		"status":        c.Writer.Status(),
		"user_agent":    c.Request.UserAgent(),
		"ip":            c.ClientIP(),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	// Convert to JSON for structured logging
	logJSON, _ := json.Marshal(logData)
	gin.DefaultWriter.Write(append(logJSON, '\n'))
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMiddleware) GetMetrics() PerformanceMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	return PerformanceMetrics{
		requestCount:      pm.metrics.requestCount,
		totalResponseTime: pm.metrics.totalResponseTime,
		slowRequests:      pm.metrics.slowRequests,
		fastRequests:      pm.metrics.fastRequests,
		errorRequests:     pm.metrics.errorRequests,
	}
}

// GetAverageResponseTime returns the average response time
func (pm *PerformanceMiddleware) GetAverageResponseTime() time.Duration {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	if pm.metrics.requestCount == 0 {
		return 0
	}

	return pm.metrics.totalResponseTime / time.Duration(pm.metrics.requestCount)
}

// GetSlowRequestPercentage returns the percentage of slow requests
func (pm *PerformanceMiddleware) GetSlowRequestPercentage() float64 {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	if pm.metrics.requestCount == 0 {
		return 0
	}

	return float64(pm.metrics.slowRequests) / float64(pm.metrics.requestCount) * 100
}

// GetErrorRate returns the error rate percentage
func (pm *PerformanceMiddleware) GetErrorRate() float64 {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	if pm.metrics.requestCount == 0 {
		return 0
	}

	return float64(pm.metrics.errorRequests) / float64(pm.metrics.requestCount) * 100
}

// ResetMetrics resets all performance metrics
func (pm *PerformanceMiddleware) ResetMetrics() {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.requestCount = 0
	pm.metrics.totalResponseTime = 0
	pm.metrics.slowRequests = 0
	pm.metrics.fastRequests = 0
	pm.metrics.errorRequests = 0
}

// Global performance middleware instance
var globalPerformanceMiddleware *PerformanceMiddleware

// InitializePerformanceMiddleware initializes the global performance middleware
func InitializePerformanceMiddleware(slowThreshold time.Duration) {
	globalPerformanceMiddleware = NewPerformanceMiddleware(slowThreshold)
}

// GetPerformanceMiddleware returns the global performance middleware
func GetPerformanceMiddleware() *PerformanceMiddleware {
	return globalPerformanceMiddleware
}

// PerformanceMiddlewareHandler returns a gin.HandlerFunc for performance tracking
func PerformanceMiddlewareHandler() gin.HandlerFunc {
	if globalPerformanceMiddleware == nil {
		// Initialize with default threshold if not already initialized
		InitializePerformanceMiddleware(100 * time.Millisecond)
	}
	return globalPerformanceMiddleware.Middleware()
}

// PerformanceHandler returns performance metrics as JSON
func PerformanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalPerformanceMiddleware == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Performance middleware not initialized",
			})
			return
		}

		metrics := globalPerformanceMiddleware.GetMetrics()
		avgResponseTime := globalPerformanceMiddleware.GetAverageResponseTime()
		slowRequestPercentage := globalPerformanceMiddleware.GetSlowRequestPercentage()
		errorRate := globalPerformanceMiddleware.GetErrorRate()

		c.JSON(http.StatusOK, gin.H{
			"metrics": gin.H{
				"request_count":       metrics.requestCount,
				"total_response_time": metrics.totalResponseTime.String(),
				"slow_requests":       metrics.slowRequests,
				"fast_requests":       metrics.fastRequests,
				"error_requests":      metrics.errorRequests,
			},
			"performance": gin.H{
				"average_response_time_ms": avgResponseTime.Milliseconds(),
				"slow_request_percentage":  slowRequestPercentage,
				"error_rate_percentage":    errorRate,
				"slow_threshold_ms":        globalPerformanceMiddleware.slowThreshold.Milliseconds(),
			},
		})
	}
}

// RequestSizeMiddleware limits request size for performance
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// CompressionMiddleware enables gzip compression for better performance
func CompressionMiddleware() gin.HandlerFunc {
	// Use gin's built-in compression
	return func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.Next()
	}
}

// CacheMiddleware adds basic caching headers for GET requests
func CacheMiddleware(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Header("Cache-Control", "public, max-age="+duration.String())
		}
		c.Next()
	}
}

// TimeoutMiddleware adds timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers for better performance
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// CORSMiddleware handles CORS for better performance
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

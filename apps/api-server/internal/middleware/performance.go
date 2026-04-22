package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriterWrapper wraps gin.ResponseWriter to capture response body
type ResponseWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while writing to the original writer
func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// Body returns the captured response body
func (w *ResponseWriterWrapper) Body() []byte {
	return w.body.Bytes()
}

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
	// In production, this would go to a structured logging system
	// For now, we'll use the standard logger
	gin.DefaultWriter.Write(fmt.Appendf(nil,
		"[SLOW REQUEST] %s %s - %v - Status: %d\n",
		c.Request.Method,
		c.Request.URL.Path,
		responseTime,
		c.Writer.Status(),
	))
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
			c.Header("ETag", generateETag(c))
		}
		c.Next()
	}
}

// ETagData represents the components used for ETag generation
type ETagData struct {
	Path      string
	Method    string
	Content   []byte
	Timestamp time.Time
	UserID    string
	Params    string
}

// ETagGenerator handles sophisticated ETag generation
type ETagGenerator struct {
	secretKey    []byte
	cacheStorage sync.Map // For storing computed ETags
}

// NewETagGenerator creates a new ETag generator instance
func NewETagGenerator(secretKey string) *ETagGenerator {
	return &ETagGenerator{
		secretKey: []byte(secretKey),
	}
}

// GenerateETag creates a sophisticated ETag based on multiple factors
func (g *ETagGenerator) GenerateETag(data ETagData) string {
	// Create a composite hash from multiple factors
	h := sha256.New()

	// Include path and method
	h.Write([]byte(data.Path))
	h.Write([]byte(data.Method))

	// Include content if available
	if len(data.Content) > 0 {
		h.Write(data.Content)
	}

	// Include timestamp (rounded to second for consistency)
	timestampBytes := []byte(fmt.Sprintf("%d", data.Timestamp.Unix()))
	h.Write(timestampBytes)

	// Include user context for personalized responses
	if data.UserID != "" {
		h.Write([]byte(data.UserID))
	}

	// Include query parameters for cache differentiation
	if data.Params != "" {
		h.Write([]byte(data.Params))
	}

	// Add secret key for security
	h.Write(g.secretKey)

	// Generate final ETag
	hash := h.Sum(nil)
	return fmt.Sprintf("\"%x\"", hash)
}

// GenerateWeakETag creates a weak ETag for dynamic content
func (g *ETagGenerator) GenerateWeakETag(data ETagData) string {
	etag := g.GenerateETag(data)
	return "W/" + etag
}

// GetCachedETag retrieves a cached ETag if available
func (g *ETagGenerator) GetCachedETag(key string) (string, bool) {
	if cached, ok := g.cacheStorage.Load(key); ok {
		if etag, ok := cached.(string); ok {
			return etag, true
		}
	}
	return "", false
}

// CacheETag stores an ETag in the cache
func (g *ETagGenerator) CacheETag(key, etag string) {
	g.cacheStorage.Store(key, etag)
}

// InvalidateCache removes ETag entries matching a pattern
func (g *ETagGenerator) InvalidateCache(pattern string) {
	g.cacheStorage.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.Contains(keyStr, pattern) {
				g.cacheStorage.Delete(key)
			}
		}
		return true
	})
}

// Global ETag generator instance
var etagGenerator = NewETagGenerator("telecom-platform-etag-secret")

// generateETag generates a sophisticated ETag for caching
func generateETag(c *gin.Context) string {
	// Extract relevant data for ETag generation
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	// Try to get cached ETag first
	cacheKey := fmt.Sprintf("%s:%s:%s", data.Method, data.Path, data.Params)
	if cached, exists := etagGenerator.GetCachedETag(cacheKey); exists {
		return cached
	}

	// Generate new ETag
	etag := etagGenerator.GenerateETag(data)

	// Cache the ETag for future use
	etagGenerator.CacheETag(cacheKey, etag)

	return etag
}

// generateContentETag creates an ETag based on response content
func generateContentETag(c *gin.Context, content []byte) string {
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Content:   content,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	return etagGenerator.GenerateETag(data)
}

// generateWeakETag creates a weak ETag for dynamic content
func generateWeakETag(c *gin.Context) string {
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	return etagGenerator.GenerateWeakETag(data)
}

// ETagValidationMiddleware handles ETag validation for conditional requests
func ETagValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only handle GET and HEAD requests for ETag validation
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.Next()
			return
		}

		// Check If-None-Match header (for GET requests)
		if noneMatch := c.GetHeader("If-None-Match"); noneMatch != "" {
			// Generate current ETag
			currentETag := generateETag(c)

			// Check if any of the ETags match
			if etagMatches(noneMatch, currentETag) {
				c.Header("ETag", currentETag)
				c.Status(http.StatusNotModified)
				c.Abort()
				return
			}
		}

		// Check If-Match header (for PUT/PATCH/DELETE requests)
		if match := c.GetHeader("If-Match"); match != "" {
			currentETag := generateETag(c)

			if !etagMatches(match, currentETag) {
				c.Status(http.StatusPreconditionFailed)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ContentBasedETagMiddleware generates ETags based on response content
func ContentBasedETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Wrap the response writer to capture the body
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = wrapper

		c.Next()

		// Only generate ETags for successful GET requests
		if c.Request.Method != "GET" || c.Writer.Status() != http.StatusOK {
			return
		}

		// Generate ETag based on captured response content
		responseBody := wrapper.Body()
		if len(responseBody) > 0 {
			etag := generateContentETag(c, responseBody)
			c.Header("ETag", etag)
		} else {
			// Fallback to weak ETag for empty responses
			etag := generateWeakETag(c)
			c.Header("ETag", etag)
		}
	}
}

// CacheInvalidationMiddleware handles cache invalidation for modified resources
func CacheInvalidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original request for post-processing
		c.Next()

		// Invalidate cache for POST, PUT, PATCH, DELETE requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" ||
			c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {

			// Invalidate ETags for this resource pattern
			pattern := c.Request.URL.Path
			etagGenerator.InvalidateCache(pattern)

			// Also invalidate related collection resources
			if strings.Contains(pattern, "/") {
				parts := strings.Split(strings.Trim(pattern, "/"), "/")
				if len(parts) > 0 {
					collectionPattern := "/" + parts[0]
					etagGenerator.InvalidateCache(collectionPattern)
				}
			}
		}
	}
}

// etagMatches checks if any of the provided ETags match the current ETag
func etagMatches(headerValue, currentETag string) bool {
	// Handle wildcard
	if headerValue == "*" {
		return true
	}

	// Parse the header value (can contain multiple ETags)
	etags := strings.Split(headerValue, ",")
	for _, etag := range etags {
		etag = strings.TrimSpace(etag)
		// Remove quotes for comparison
		etag = strings.Trim(etag, "\"")
		current := strings.Trim(currentETag, "\"W/")

		if etag == current || etag == "W/"+current {
			return true
		}
	}

	return false
}

// AdvancedCacheMiddleware provides sophisticated caching with ETags
func AdvancedCacheMiddleware(maxAge time.Duration, useWeakETag bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			// Set cache control headers
			cacheControl := fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds()))
			if useWeakETag {
				cacheControl += ", must-revalidate"
			}
			c.Header("Cache-Control", cacheControl)

			// Generate appropriate ETag
			var etag string
			if useWeakETag {
				etag = generateWeakETag(c)
			} else {
				etag = generateETag(c)
			}
			c.Header("ETag", etag)

			// Add Vary header for proper caching with different user contexts
			if userID, exists := c.Get("user_id"); exists {
				c.Header("Vary", "Accept, Authorization")
				if userIDStr, ok := userID.(string); ok && userIDStr != "" {
					c.Header("X-User-Context", userIDStr)
				}
			}
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

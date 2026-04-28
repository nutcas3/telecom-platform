package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple rate limiter implementation using in-memory storage
type SimpleRateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewSimpleRateLimiter creates a new simple rate limiter
func NewSimpleRateLimiter(limit int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed based on the key
func (r *SimpleRateLimiter) Allow(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()

	// Clean old requests
	if requests, exists := r.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < r.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		r.requests[key] = validRequests
	}

	// Check if under limit
	if len(r.requests[key]) < r.limit {
		r.requests[key] = append(r.requests[key], now)
		return true
	}

	return false
}

// Global rate limiter instance
var (
	authRateLimiter  = NewSimpleRateLimiter(10, time.Minute)  // 10 requests per minute for auth
	apiRateLimiter   = NewSimpleRateLimiter(100, time.Minute) // 100 requests per minute for general API
	readRateLimiter  = NewSimpleRateLimiter(200, time.Minute) // 200 requests per minute for read endpoints
	adminRateLimiter = NewSimpleRateLimiter(50, time.Minute)  // 50 requests per minute for admin endpoints
)

// RateLimit creates a rate limiting middleware
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(requestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByEndpoint creates different rate limits for different endpoint types
func RateLimitByEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		key := c.ClientIP()

		var allowed bool

		switch {
		case path == "/v1/auth/login" || path == "/v1/auth/register":
			// Strict rate limiting for auth endpoints
			allowed = authRateLimiter.Allow(key)
		case method == "GET" && (path == "/v1/services" || path == "/v1/monitoring/metrics"):
			// Lenient rate limiting for read-heavy endpoints
			allowed = readRateLimiter.Allow(key)
		case path == "/v1/users" || path == "/v1/config":
			// Admin rate limiting for sensitive endpoints
			allowed = adminRateLimiter.Allow(key)
		default:
			// Default rate limiting for other endpoints
			allowed = apiRateLimiter.Allow(key)
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser creates rate limiting based on authenticated user
func RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(requestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		var key string

		// Try to get user ID from context
		if userID, exists := c.Get("user_id"); exists {
			key = "user:" + userID.(string)
		} else {
			// Fallback to IP if not authenticated
			key = "anon:" + c.ClientIP()
		}

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByIP creates rate limiting based on IP address only
func RateLimitByIP(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(requestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithHeaders adds rate limit headers to responses
func RateLimitWithHeaders(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(requestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		key := c.ClientIP()

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
		c.Header("X-RateLimit-Window", "60")

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

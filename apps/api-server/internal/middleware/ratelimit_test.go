package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSimpleRateLimiter_Allow(t *testing.T) {
	limiter := NewSimpleRateLimiter(5, time.Second)

	// Should allow first 5 requests
	for i := range 5 {
		assert.True(t, limiter.Allow("test-key"), "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	assert.False(t, limiter.Allow("test-key"), "6th request should be denied")
}

func TestSimpleRateLimiter_WindowReset(t *testing.T) {
	limiter := NewSimpleRateLimiter(2, 100*time.Millisecond)

	// Use up the limit
	assert.True(t, limiter.Allow("test-key"))
	assert.True(t, limiter.Allow("test-key"))
	assert.False(t, limiter.Allow("test-key"))

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	assert.True(t, limiter.Allow("test-key"))
}

func TestSimpleRateLimiter_DifferentKeys(t *testing.T) {
	limiter := NewSimpleRateLimiter(2, time.Second)

	// Different keys should have independent limits
	assert.True(t, limiter.Allow("key1"))
	assert.True(t, limiter.Allow("key1"))
	assert.False(t, limiter.Allow("key1"))

	// key2 should still have its full limit
	assert.True(t, limiter.Allow("key2"))
	assert.True(t, limiter.Allow("key2"))
}

func TestRateLimit_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test handler
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	// Apply rate limiting middleware (10 requests per minute)
	router := gin.New()
	router.Use(RateLimit(10))
	router.GET("/test", handler)

	// Make 11 requests
	for i := range 10 {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 11th request should be rate limited
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Rate limit exceeded", response["error"])
}

func TestRateLimitByEndpoint_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router := gin.New()
	router.Use(RateLimitByEndpoint())
	router.GET("/v1/auth/login", handler)
	router.GET("/v1/services", handler)
	router.GET("/v1/users", handler)
	router.GET("/v1/other", handler)

	// Test auth endpoint (strict limit)
	for i := range 11 {
		req, _ := http.NewRequest("GET", "/v1/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 10 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimitByUser_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router := gin.New()
	router.Use(RateLimitByUser(5))
	router.GET("/test", handler)

	// Test without user ID (should use IP-based limiting)
	for i := range 6 {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 5 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimitByIP_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router := gin.New()
	router.Use(RateLimitByIP(3))
	router.GET("/test", handler)

	// Test IP-based limiting
	for i := range 4 {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 3 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimitWithHeaders_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	router := gin.New()
	router.Use(RateLimitWithHeaders(10))
	router.GET("/test", handler)

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check for rate limit headers
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "60", w.Header().Get("X-RateLimit-Window"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with very high limit (effectively no limit)
	router := gin.New()
	router.Use(RateLimit(10000)) // Very high limit to simulate no rate limiting
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Should work without hitting rate limit
	for range 20 {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimit_ConcurrentAccess(t *testing.T) {
	limiter := NewSimpleRateLimiter(100, time.Second)

	// Test concurrent access
	done := make(chan bool, 200)

	// Launch 200 goroutines, each making 1 request
	for range 200 {
		go func() {
			allowed := limiter.Allow("concurrent-test")
			done <- allowed
		}()
	}

	// Count allowed requests
	allowedCount := 0
	for range 200 {
		if <-done {
			allowedCount++
		}
	}

	// Should allow exactly 100 requests
	assert.Equal(t, 100, allowedCount)
}

func TestRateLimit_CleanupOldRequests(t *testing.T) {
	limiter := NewSimpleRateLimiter(5, 100*time.Millisecond)

	// Use up the limit
	for range 5 {
		assert.True(t, limiter.Allow("cleanup-test"))
	}
	assert.False(t, limiter.Allow("cleanup-test"))

	// Wait for old requests to be cleaned up
	time.Sleep(200 * time.Millisecond)

	// Should be able to make requests again
	assert.True(t, limiter.Allow("cleanup-test"))

	// Check that old requests were cleaned up (internal state should be smaller)
	// This is more of an implementation detail test, but ensures cleanup works
}

// Benchmark tests
func BenchmarkSimpleRateLimiter_Allow(b *testing.B) {
	limiter := NewSimpleRateLimiter(1000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("benchmark-key")
	}
}

func BenchmarkRateLimit_Middleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit(1000))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

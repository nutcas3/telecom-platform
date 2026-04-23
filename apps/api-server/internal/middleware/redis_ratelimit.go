package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
	prefix string
	limit  int64
	window time.Duration
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool
	Remaining  int64
	ResetTime  time.Time
	TotalLimit int64
	RetryAfter time.Duration
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, prefix string, limit int64, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		prefix: prefix,
		limit:  limit,
		window: window,
	}
}

// Allow checks if a request is allowed using sliding window algorithm
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (*RateLimitResult, error) {
	now := time.Now()
	redisKey := fmt.Sprintf("%s:%s", r.prefix, key)

	// Use Lua script for atomic sliding window operations
	luaScript := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_start = now - window
		
		-- Remove expired entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Count current requests
		local current = redis.call('ZCARD', key)
		
		-- Check if under limit
		if current < limit then
			-- Add current request
			redis.call('ZADD', key, now, now)
			-- Set expiration
			redis.call('EXPIRE', key, math.ceil(window))
			return {1, limit - current - 1, now + window}
		else
			-- Get oldest request time for retry-after
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local retry_after = 0
			if #oldest > 0 then
				retry_after = oldest[2] + window - now
			end
			return {0, 0, now + window, retry_after}
		end
	`

	result, err := r.client.Eval(ctx, luaScript, []string{redisKey},
		now.Unix(), r.window.Seconds(), r.limit).Result()

	if err != nil {
		return nil, fmt.Errorf("redis rate limit check failed: %w", err)
	}

	// Parse Lua script result
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return nil, fmt.Errorf("invalid redis result format")
	}

	allowed, _ := resultSlice[0].(int64)
	remaining, _ := resultSlice[1].(int64)
	resetTimeUnix, _ := resultSlice[2].(int64)
	resetTime := time.Unix(resetTimeUnix, 0)

	retryAfter := time.Duration(0)
	if len(resultSlice) >= 4 {
		retryAfterSeconds, _ := resultSlice[3].(int64)
		retryAfter = time.Duration(retryAfterSeconds) * time.Second
	}

	return &RateLimitResult{
		Allowed:    allowed == 1,
		Remaining:  remaining,
		ResetTime:  resetTime,
		TotalLimit: r.limit,
		RetryAfter: retryAfter,
	}, nil
}

// AllowTokenBucket implements token bucket algorithm using Redis
func (r *RedisRateLimiter) AllowTokenBucket(ctx context.Context, key string) (*RateLimitResult, error) {
	now := time.Now()
	redisKey := fmt.Sprintf("%s:bucket:%s", r.prefix, key)

	// Token bucket Lua script
	luaScript := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local refill_rate = tonumber(ARGV[3])
		local refill_interval = tonumber(ARGV[4])
		
		-- Get current bucket state
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or capacity
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add
		local time_passed = now - last_refill
		local tokens_to_add = math.floor((time_passed / refill_interval) * refill_rate)
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		-- Check if request can be allowed
		if tokens >= 1 then
			tokens = tokens - 1
			-- Update bucket state
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, math.ceil(refill_interval * capacity / refill_rate))
			return {1, tokens, now + refill_interval}
		else
			-- Calculate retry after
			local retry_after = math.ceil((1 - tokens) * refill_interval / refill_rate)
			return {0, tokens, now + refill_interval, retry_after}
		end
	`

	// Calculate refill rate (tokens per second)
	refillRate := float64(r.limit) / r.window.Seconds()
	refillInterval := 1.0 // 1 second intervals

	result, err := r.client.Eval(ctx, luaScript, []string{redisKey},
		now.Unix(), r.limit, refillRate, refillInterval).Result()

	if err != nil {
		return nil, fmt.Errorf("redis token bucket check failed: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return nil, fmt.Errorf("invalid redis result format")
	}

	allowed, _ := resultSlice[0].(int64)
	remaining, _ := resultSlice[1].(int64)
	resetTimeUnix, _ := resultSlice[2].(int64)
	resetTime := time.Unix(resetTimeUnix, 0)

	retryAfter := time.Duration(0)
	if len(resultSlice) >= 4 {
		retryAfterSeconds, _ := resultSlice[3].(int64)
		retryAfter = time.Duration(retryAfterSeconds) * time.Second
	}

	return &RateLimitResult{
		Allowed:    allowed == 1,
		Remaining:  remaining,
		ResetTime:  resetTime,
		TotalLimit: r.limit,
		RetryAfter: retryAfter,
	}, nil
}

// Reset clears the rate limit for a specific key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("%s:%s", r.prefix, key)
	return r.client.Del(ctx, redisKey).Err()
}

// GetStats returns current rate limit statistics for a key
func (r *RedisRateLimiter) GetStats(ctx context.Context, key string) (*RateLimitResult, error) {
	now := time.Now()
	windowStart := now.Add(-r.window)
	redisKey := fmt.Sprintf("%s:%s", r.prefix, key)

	// Count requests in current window
	count, err := r.client.ZCount(ctx, redisKey, fmt.Sprintf("%d", windowStart.Unix()), fmt.Sprintf("%d", now.Unix())).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit stats: %w", err)
	}

	remaining := r.limit - int64(count)
	if remaining < 0 {
		remaining = 0
	}

	return &RateLimitResult{
		Allowed:    remaining > 0,
		Remaining:  remaining,
		ResetTime:  now.Add(r.window),
		TotalLimit: r.limit,
		RetryAfter: time.Duration(0),
	}, nil
}

// RedisRateLimitMiddleware creates a Gin middleware for Redis-based rate limiting
func RedisRateLimitMiddleware(limiter *RedisRateLimiter, keyExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyExtractor(c)
		if key == "" {
			c.Next()
			return
		}

		result, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			// Log error but allow request to proceed
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(result.TotalLimit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			if result.RetryAfter > 0 {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"details": gin.H{
					"limit":      result.TotalLimit,
					"remaining":  result.Remaining,
					"reset_time": result.ResetTime,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RedisTokenBucketMiddleware creates a Gin middleware for Redis token bucket rate limiting
func RedisTokenBucketMiddleware(limiter *RedisRateLimiter, keyExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyExtractor(c)
		if key == "" {
			c.Next()
			return
		}

		result, err := limiter.AllowTokenBucket(c.Request.Context(), key)
		if err != nil {
			// Log error but allow request to proceed
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(result.TotalLimit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			if result.RetryAfter > 0 {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"details": gin.H{
					"limit":      result.TotalLimit,
					"remaining":  result.Remaining,
					"reset_time": result.ResetTime,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Key extractor functions

// IPKeyExtractor extracts rate limit key from client IP
func IPKeyExtractor(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyExtractor extracts rate limit key from authenticated user ID
func UserKeyExtractor(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return "user:" + id
		}
	}
	return IPKeyExtractor(c)
}

// APIKeyExtractor extracts rate limit key from API key
func APIKeyExtractor(c *gin.Context) string {
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return "api:" + apiKey
	}
	return UserKeyExtractor(c)
}

// EndpointKeyExtractor creates a key extractor that includes the endpoint
func EndpointKeyExtractor(baseExtractor func(*gin.Context) string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		baseKey := baseExtractor(c)
		return fmt.Sprintf("%s:%s:%s", baseKey, c.Request.Method, c.FullPath())
	}
}

// MultiTierRateLimiter implements multiple rate limit tiers
type MultiTierRateLimiter struct {
	limiters map[string]*RedisRateLimiter
}

// NewMultiTierRateLimiter creates a multi-tier rate limiter
func NewMultiTierRateLimiter(client *redis.Client, tiers map[string]struct {
	Limit  int64
	Window time.Duration
	Prefix string
}) *MultiTierRateLimiter {
	limiters := make(map[string]*RedisRateLimiter)

	for name, config := range tiers {
		limiters[name] = NewRedisRateLimiter(client, config.Prefix, config.Limit, config.Window)
	}

	return &MultiTierRateLimiter{limiters: limiters}
}

// CheckTier checks rate limit for a specific tier
func (m *MultiTierRateLimiter) CheckTier(ctx context.Context, tier, key string) (*RateLimitResult, error) {
	limiter, exists := m.limiters[tier]
	if !exists {
		return nil, fmt.Errorf("rate limit tier '%s' not found", tier)
	}

	return limiter.Allow(ctx, key)
}

// MultiTierMiddleware creates middleware for multi-tier rate limiting
func MultiTierMiddleware(multiLimiter *MultiTierRateLimiter, tierExtractor func(*gin.Context) string, keyExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tier := tierExtractor(c)
		key := keyExtractor(c)

		if tier == "" || key == "" {
			c.Next()
			return
		}

		result, err := multiLimiter.CheckTier(c.Request.Context(), tier, key)
		if err != nil {
			c.Next()
			return
		}

		// Set rate limit headers with tier information
		c.Header("X-RateLimit-Limit", strconv.FormatInt(result.TotalLimit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		c.Header("X-RateLimit-Tier", tier)

		if !result.Allowed {
			if result.RetryAfter > 0 {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"details": gin.H{
					"tier":       tier,
					"limit":      result.TotalLimit,
					"remaining":  result.Remaining,
					"reset_time": result.ResetTime,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

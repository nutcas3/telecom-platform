package tenant

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var ErrTenantNotFound = errors.New("tenant not found")

type TenantRateLimiter struct {
	planLimits map[TenantPlan]RateLimit
	tenantUsage map[string]*TenantUsageTracker
	mu sync.RWMutex
	logger *logrus.Logger
}

type RateLimit struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
}

type TenantUsageTracker struct {
	TenantID string
	Plan     TenantPlan

	minuteRequests *SlidingWindowCounter
	hourRequests   *SlidingWindowCounter
	dayRequests    *SlidingWindowCounter

	// Last reset time
	lastReset time.Time
}

type SlidingWindowCounter struct {
	window   time.Duration
	maxCount int
	requests []time.Time
	mu       sync.Mutex
}

func NewTenantRateLimiter(logger *logrus.Logger) *TenantRateLimiter {
	return &TenantRateLimiter{
		planLimits: map[TenantPlan]RateLimit{
			TenantPlanFree: {
				RequestsPerMinute: 60,
				RequestsPerHour:   1000,
				RequestsPerDay:    10000,
				BurstSize:         10,
			},
			TenantPlanBasic: {
				RequestsPerMinute: 120,
				RequestsPerHour:   5000,
				RequestsPerDay:    50000,
				BurstSize:         20,
			},
			TenantPlanPro: {
				RequestsPerMinute: 300,
				RequestsPerHour:   15000,
				RequestsPerDay:    150000,
				BurstSize:         50,
			},
			TenantPlanEnterprise: {
				RequestsPerMinute: 1000,
				RequestsPerHour:   100000,
				RequestsPerDay:    1000000,
				BurstSize:         100,
			},
		},
		tenantUsage: make(map[string]*TenantUsageTracker),
		logger:      logger,
	}
}

func (trl *TenantRateLimiter) AllowRequest(ctx context.Context, tenantID string, plan TenantPlan) bool {
	trl.mu.Lock()
	defer trl.mu.Unlock()

	// Get or create usage tracker for tenant
	tracker, exists := trl.tenantUsage[tenantID]
	if !exists {
		tracker = &TenantUsageTracker{
			TenantID:  tenantID,
			Plan:      plan,
			lastReset: time.Now(),
		}

		// Initialize sliding windows
		limits := trl.planLimits[plan]
		tracker.minuteRequests = NewSlidingWindowCounter(time.Minute, limits.RequestsPerMinute)
		tracker.hourRequests = NewSlidingWindowCounter(time.Hour, limits.RequestsPerHour)
		tracker.dayRequests = NewSlidingWindowCounter(24*time.Hour, limits.RequestsPerDay)

		trl.tenantUsage[tenantID] = tracker
	}

	// Check all rate limits
	limits := trl.planLimits[plan]

	if !tracker.minuteRequests.Allow(limits.BurstSize) {
		trl.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"plan":      plan,
			"limit":     "minute",
			"current":   tracker.minuteRequests.Count(),
			"max":       limits.RequestsPerMinute,
		}).Warn("Rate limit exceeded: minute")
		return false
	}

	if !tracker.hourRequests.Allow(limits.BurstSize) {
		trl.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"plan":      plan,
			"limit":     "hour",
			"current":   tracker.hourRequests.Count(),
			"max":       limits.RequestsPerHour,
		}).Warn("Rate limit exceeded: hour")
		return false
	}

	if !tracker.dayRequests.Allow(limits.BurstSize) {
		trl.logger.WithFields(logrus.Fields{
			"tenant_id": tenantID,
			"plan":      plan,
			"limit":     "day",
			"current":   tracker.dayRequests.Count(),
			"max":       limits.RequestsPerDay,
		}).Warn("Rate limit exceeded: day")
		return false
	}

	return true
}

// GetUsageStats returns current usage statistics for a tenant
func (trl *TenantRateLimiter) GetUsageStats(tenantID string) (*RateLimitUsage, error) {
	trl.mu.RLock()
	defer trl.mu.RUnlock()

	tracker, exists := trl.tenantUsage[tenantID]
	if !exists {
		return nil, ErrTenantNotFound
	}

	limits := trl.planLimits[tracker.Plan]

	return &RateLimitUsage{
		TenantID:        tenantID,
		Plan:            tracker.Plan,
		MinuteRequests:  tracker.minuteRequests.Count(),
		MinuteLimit:     limits.RequestsPerMinute,
		HourRequests:    tracker.hourRequests.Count(),
		HourLimit:       limits.RequestsPerHour,
		DayRequests:     tracker.dayRequests.Count(),
		DayLimit:        limits.RequestsPerDay,
		RemainingMinute: limits.RequestsPerMinute - tracker.minuteRequests.Count(),
		RemainingHour:   limits.RequestsPerHour - tracker.hourRequests.Count(),
		RemainingDay:    limits.RequestsPerDay - tracker.dayRequests.Count(),
		ResetTime:       tracker.lastReset.Add(24 * time.Hour),
	}, nil
}

// UpdateTenantPlan updates the rate limiting plan for a tenant
func (trl *TenantRateLimiter) UpdateTenantPlan(tenantID string, newPlan TenantPlan) {
	trl.mu.Lock()
	defer trl.mu.Unlock()

	tracker, exists := trl.tenantUsage[tenantID]
	if !exists {
		return
	}

	// Update plan and reinitialize counters with new limits
	tracker.Plan = newPlan
	limits := trl.planLimits[newPlan]

	tracker.minuteRequests = NewSlidingWindowCounter(time.Minute, limits.RequestsPerMinute)
	tracker.hourRequests = NewSlidingWindowCounter(time.Hour, limits.RequestsPerHour)
	tracker.dayRequests = NewSlidingWindowCounter(24*time.Hour, limits.RequestsPerDay)

	trl.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"old_plan":  tracker.Plan,
		"new_plan":  newPlan,
	}).Info("Updated tenant rate limiting plan")
}

// CleanupExpiredTrackers removes inactive tenant trackers
func (trl *TenantRateLimiter) CleanupExpiredTrackers() {
	trl.mu.Lock()
	defer trl.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)

	for tenantID, tracker := range trl.tenantUsage {
		if tracker.lastReset.Before(cutoff) {
			delete(trl.tenantUsage, tenantID)
			trl.logger.WithField("tenant_id", tenantID).Info("Cleaned up expired rate limit tracker")
		}
	}
}

// NewSlidingWindowCounter creates a new sliding window counter
func NewSlidingWindowCounter(window time.Duration, maxCount int) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		window:   window,
		maxCount: maxCount,
		requests: make([]time.Time, 0),
	}
}

// Allow checks if a request should be allowed and records it
func (swc *SlidingWindowCounter) Allow(burstSize int) bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()

	// Remove expired requests
	cutoff := now.Add(-swc.window)
	validRequests := make([]time.Time, 0)
	for _, req := range swc.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	swc.requests = validRequests

	// Check if we can allow this request
	if len(swc.requests) >= swc.maxCount {
		return false
	}

	// Allow the request
	swc.requests = append(swc.requests, now)
	return true
}

// Count returns the current count of requests in the window
func (swc *SlidingWindowCounter) Count() int {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swc.window)

	count := 0
	for _, req := range swc.requests {
		if req.After(cutoff) {
			count++
		}
	}

	return count
}

// RateLimitUsage represents current rate limit usage for a tenant
type RateLimitUsage struct {
	TenantID        string     `json:"tenant_id"`
	Plan            TenantPlan `json:"plan"`
	MinuteRequests  int        `json:"minute_requests"`
	MinuteLimit     int        `json:"minute_limit"`
	HourRequests    int        `json:"hour_requests"`
	HourLimit       int        `json:"hour_limit"`
	DayRequests     int        `json:"day_requests"`
	DayLimit        int        `json:"day_limit"`
	RemainingMinute int        `json:"remaining_minute"`
	RemainingHour   int        `json:"remaining_hour"`
	RemainingDay    int        `json:"remaining_day"`
	ResetTime       time.Time  `json:"reset_time"`
}

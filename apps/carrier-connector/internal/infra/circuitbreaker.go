package infra

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	state         CircuitState
	failureCount  int
	successCount  int
	lastFailure   time.Time
	config        CircuitBreakerConfig
	mu            sync.RWMutex
	onStateChange func(name string, from, to CircuitState)
}

// CircuitBreakerConfig configures the circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	HalfOpenMaxCalls int
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:   name,
		state:  StateClosed,
		config: config,
	}
}

// ErrCircuitOpen is returned when the circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs the function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// ExecuteWithFallback runs with fallback on circuit open
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	if !cb.allowRequest() {
		if fallback != nil {
			return fallback()
		}
		return ErrCircuitOpen
	}

	err := fn()
	cb.recordResult(err)
	if err != nil && fallback != nil {
		return fallback()
	}
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.config.Timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.transitionTo(StateHalfOpen)
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		cb.transitionTo(StateOpen)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++

	switch cb.state {
	case StateHalfOpen:
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
		}
	case StateClosed:
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) transitionTo(state CircuitState) {
	if cb.state == state {
		return
	}
	oldState := cb.state
	cb.state = state
	cb.failureCount = 0
	cb.successCount = 0

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, state)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		Name:         cb.name,
		State:        cb.state.String(),
		FailureCount: cb.failureCount,
		SuccessCount: cb.successCount,
		LastFailure:  cb.lastFailure,
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	Name         string    `json:"name"`
	State        string    `json:"state"`
	FailureCount int       `json:"failure_count"`
	SuccessCount int       `json:"success_count"`
	LastFailure  time.Time `json:"last_failure"`
}

// OnStateChange sets the state change callback
func (cb *CircuitBreaker) OnStateChange(fn func(name string, from, to CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

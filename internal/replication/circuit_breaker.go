package replication

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the circuit breaker state
type CircuitState string

const (
	StateClosed   CircuitState = "closed"    // Normal operation
	StateOpen     CircuitState = "open"      // Failing, rejecting requests
	StateHalfOpen CircuitState = "half_open" // Testing if service recovered
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening
	MaxFailures int

	// Timeout is how long to wait before attempting to close an open circuit
	Timeout time.Duration

	// HalfOpenMaxAttempts is max successful requests in half-open before closing
	HalfOpenMaxAttempts int
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		HalfOpenMaxAttempts: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern for replication
type CircuitBreaker struct {
	config CircuitBreakerConfig

	mu              sync.RWMutex
	state           CircuitState
	failures        int
	successes       int
	lastFailureTime time.Time
	lastStateChange time.Time
	totalFailures   int64
	totalSuccesses  int64
	totalRejections int64
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check if we should allow the call
	if !cb.canAttempt() {
		cb.recordRejection()
		return fmt.Errorf("circuit breaker is %s, rejecting call", cb.GetState())
	}

	// Execute the function
	err := fn()

	// Record the result
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// canAttempt checks if we should allow an attempt
func (cb *CircuitBreaker) canAttempt() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastStateChange) > cb.config.Timeout {
			// Try transitioning to half-open
			// This will be done in recordAttempt
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordSuccess records a successful call
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccesses++
	cb.failures = 0

	switch cb.state {
	case StateClosed:
		// Stay closed
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.HalfOpenMaxAttempts {
			cb.transitionTo(StateClosed)
		}
	case StateOpen:
		// This shouldn't happen, but transition to half-open
		cb.transitionTo(StateHalfOpen)
	}
}

// recordFailure records a failed call
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Any failure in half-open immediately opens the circuit
		cb.transitionTo(StateOpen)
	case StateOpen:
		// Stay open, reset timeout
		cb.lastStateChange = time.Now()
	}
}

// recordRejection records a rejected call
func (cb *CircuitBreaker) recordRejection() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRejections++

	// If we're open and timeout has passed, try half-open
	if cb.state == StateOpen && time.Since(cb.lastStateChange) > cb.config.Timeout {
		cb.transitionTo(StateHalfOpen)
	}
}

// transitionTo transitions to a new state
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	cb.failures = 0
	cb.successes = 0

	// Log state change
	// In production, use monitoring.Log
	_ = oldState // Avoid unused variable
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:             cb.state,
		Failures:          cb.failures,
		Successes:         cb.successes,
		TotalFailures:     cb.totalFailures,
		TotalSuccesses:    cb.totalSuccesses,
		TotalRejections:   cb.totalRejections,
		LastFailureTime:   cb.lastFailureTime,
		LastStateChange:   cb.lastStateChange,
		TimeSinceLastFail: time.Since(cb.lastFailureTime),
		TimeSinceStateChg: time.Since(cb.lastStateChange),
	}
}

// CircuitBreakerStats holds statistics about the circuit breaker
type CircuitBreakerStats struct {
	State             CircuitState
	Failures          int
	Successes         int
	TotalFailures     int64
	TotalSuccesses    int64
	TotalRejections   int64
	LastFailureTime   time.Time
	LastStateChange   time.Time
	TimeSinceLastFail time.Duration
	TimeSinceStateChg time.Duration
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastStateChange = time.Now()
}

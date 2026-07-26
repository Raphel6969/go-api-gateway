package breaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOption = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Aloow check if a request is permitted to pass
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	state := cb.state
	last := cb.lastFailure
	cb.mu.RUnlock()

	if state == StateClosed {
		return true
	}

	if state == StateOpen {
		if time.Since(last) > cb.resetTimeout {
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.mu.Unlock()
			return true
		}
		return false
	}

	return state == StateHalfOpen
}

func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.failures = 0
		cb.state = StateClosed
		return
	}

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

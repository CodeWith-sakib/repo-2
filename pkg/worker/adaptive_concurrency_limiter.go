package worker

import (
	"math"
	"sync"
	"time"
)

// AdaptiveConcurrencyLimiter adjusts worker concurrency dynamically using TCP Vegas/Vegas-style RTT gradient.
type AdaptiveConcurrencyLimiter struct {
	mu           sync.RWMutex
	currentLimit float64
	minLimit     float64
	maxLimit     float64
	rttMin       time.Duration
	rttCurrent   time.Duration
	smoothing    float64
	inFlight     int
}

// NewAdaptiveConcurrencyLimiter creates an adaptive limiter with bounds.
func NewAdaptiveConcurrencyLimiter(initial, min, max float64) *AdaptiveConcurrencyLimiter {
	if initial <= 0 {
		initial = 10.0
	}
	if min <= 0 {
		min = 2.0
	}
	if max <= 0 {
		max = 100.0
	}
	return &AdaptiveConcurrencyLimiter{
		currentLimit: initial,
		minLimit:     min,
		maxLimit:     max,
		smoothing:    0.2,
	}
}

// TryAcquire checks if an additional concurrent worker slot is available.
func (acl *AdaptiveConcurrencyLimiter) TryAcquire() bool {
	acl.mu.Lock()
	defer acl.mu.Unlock()

	if float64(acl.inFlight) < acl.currentLimit {
		acl.inFlight++
		return true
	}
	return false
}

// Release registers task completion latency and dynamically tunes the concurrency limit.
func (acl *AdaptiveConcurrencyLimiter) Release(rtt time.Duration) {
	acl.mu.Lock()
	defer acl.mu.Unlock()

	if acl.inFlight > 0 {
		acl.inFlight--
	}

	if rtt <= 0 {
		return
	}

	if acl.rttMin == 0 || rtt < acl.rttMin {
		acl.rttMin = rtt
	}

	if acl.rttCurrent == 0 {
		acl.rttCurrent = rtt
	} else {
		acl.rttCurrent = time.Duration(float64(acl.rttCurrent)*(1.0-acl.smoothing) + float64(rtt)*acl.smoothing)
	}

	// Gradient: rttMin / rttCurrent
	gradient := float64(acl.rttMin) / float64(acl.rttCurrent)
	if gradient > 1.0 {
		// Low latency / healthy system: expand concurrency
		acl.currentLimit = math.Min(acl.maxLimit, acl.currentLimit+0.5)
	} else if gradient < 0.8 {
		// Backpressure detected: throttle concurrency
		acl.currentLimit = math.Max(acl.minLimit, acl.currentLimit*gradient)
	}
}

// Limit returns current concurrency ceiling.
func (acl *AdaptiveConcurrencyLimiter) Limit() int {
	acl.mu.RLock()
	defer acl.mu.RUnlock()
	return int(math.Round(acl.currentLimit))
}

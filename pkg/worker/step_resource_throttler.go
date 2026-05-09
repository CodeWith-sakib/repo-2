package worker

import (
	"context"
	"sync"
	"time"
)

// StepResourceThrottler throttles task scheduling when system metrics (CPU load, memory) exceed thresholds.
type StepResourceThrottler struct {
	mu          sync.RWMutex
	isThrottled bool
	maxLoadAvg  float64
	cooldown    time.Duration
	lastToggled time.Time
}

// NewStepResourceThrottler creates a system resource throttler.
func NewStepResourceThrottler(maxLoad float64, cooldown time.Duration) *StepResourceThrottler {
	if maxLoad <= 0 {
		maxLoad = 8.0
	}
	if cooldown <= 0 {
		cooldown = 10 * time.Second
	}
	return &StepResourceThrottler{
		maxLoadAvg: maxLoad,
		cooldown:   cooldown,
	}
}

// UpdateSystemLoad evaluates observed system load and toggles throttle state.
func (t *StepResourceThrottler) UpdateSystemLoad(currentLoad float64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now().UTC()
	if !t.lastToggled.IsZero() && now.Sub(t.lastToggled) < t.cooldown {
		return t.isThrottled
	}

	if currentLoad >= t.maxLoadAvg && !t.isThrottled {
		t.isThrottled = true
		t.lastToggled = now
	} else if currentLoad < t.maxLoadAvg*0.7 && t.isThrottled {
		// Hysteresis release at 70% threshold
		t.isThrottled = false
		t.lastToggled = now
	}

	return t.isThrottled
}

// ShouldThrottle returns current throttled state.
func (t *StepResourceThrottler) ShouldThrottle() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isThrottled
}

// WaitIfThrottled pauses execution briefly if throttler is active.
func (t *StepResourceThrottler) WaitIfThrottled(ctx context.Context, pause time.Duration) error {
	if !t.ShouldThrottle() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(pause):
		return nil
	}
}

package events

import (
	"testing"
	"time"
)

func TestEventBackpressureRegulator(t *testing.T) {
	reg := NewEventBackpressureRegulator(1000, 100*time.Millisecond)

	// Healthy backlog (200 / 1000) -> 0 delay
	reg.UpdateQueueDepth(200)
	if reg.ComputeThrottleDelay() != 0 {
		t.Errorf("expected 0 throttle delay under low load, got %v", reg.ComputeThrottleDelay())
	}
	if reg.IsOverloaded() {
		t.Error("expected not overloaded")
	}

	// High backlog (800 / 1000) -> positive throttle delay
	reg.UpdateQueueDepth(800)
	delay := reg.ComputeThrottleDelay()
	if delay < 50*time.Millisecond {
		t.Errorf("expected significant throttle delay under high load, got %v", delay)
	}

	// Saturated backlog (1200 / 1000) -> overloaded
	reg.UpdateQueueDepth(1200)
	if !reg.IsOverloaded() {
		t.Error("expected overloaded under saturated load")
	}
}

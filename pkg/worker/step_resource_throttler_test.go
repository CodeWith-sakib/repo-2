package worker

import (
	"context"
	"testing"
	"time"
)

func TestStepResourceThrottler(t *testing.T) {
	throttler := NewStepResourceThrottler(10.0, 50*time.Millisecond)

	if throttler.ShouldThrottle() {
		t.Fatal("expected initially unthrottled")
	}

	// High load -> activates throttle
	throttler.UpdateSystemLoad(12.5)
	if !throttler.ShouldThrottle() {
		t.Error("expected throttler to activate on high load")
	}

	// Test pause execution
	ctx := context.Background()
	start := time.Now()
	_ = throttler.WaitIfThrottled(ctx, 20*time.Millisecond)
	if time.Since(start) < 15*time.Millisecond {
		t.Error("expected pause when throttled")
	}

	// Cooldown pass -> release below hysteresis threshold (7.0)
	time.Sleep(60 * time.Millisecond)
	throttler.UpdateSystemLoad(5.0)
	if throttler.ShouldThrottle() {
		t.Error("expected throttler to deactivate after load drops below hysteresis")
	}
}

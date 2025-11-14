package worker

import (
	"testing"
	"time"
)

func TestVegasConcurrencyController_Adjustment(t *testing.T) {
	cfg := VegasConfig{
		MinLimit:  2,
		MaxLimit:  20,
		Alpha:     2.0,
		Beta:      4.0,
		Smoothing: 1.0, // instant smoothing for test
	}
	ctrl := NewVegasConcurrencyController(cfg)

	initialLimit := ctrl.Limit()

	// Feed low baseline latency (10ms) repeatedly -> should increase limit (queue is small)
	for i := 0; i < 5; i++ {
		ctrl.TryAcquire()
		ctrl.Release(10 * time.Millisecond)
	}

	if ctrl.Limit() <= initialLimit {
		t.Errorf("expected limit to increase from %d on fast responses, got %d", initialLimit, ctrl.Limit())
	}

	highLimit := ctrl.Limit()

	// Now feed high latency spike (100ms vs 10ms baseline) -> queue builds up -> should decrease limit
	for i := 0; i < 10; i++ {
		ctrl.TryAcquire()
		ctrl.Release(100 * time.Millisecond)
	}

	if ctrl.Limit() >= highLimit {
		t.Errorf("expected limit to decrease from %d on latency spike, got %d", highLimit, ctrl.Limit())
	}
}

func TestVegasConcurrencyController_TryAcquireCeiling(t *testing.T) {
	cfg := VegasConfig{
		MinLimit: 2,
		MaxLimit: 2,
	}
	ctrl := NewVegasConcurrencyController(cfg)

	if !ctrl.TryAcquire() {
		t.Fatal("expected 1st acquire to succeed")
	}
	if !ctrl.TryAcquire() {
		t.Fatal("expected 2nd acquire to succeed")
	}
	// 3rd acquire should fail (ceiling = 2)
	if ctrl.TryAcquire() {
		t.Error("expected 3rd acquire to fail at ceiling")
	}

	ctrl.Release(10 * time.Millisecond)
	// Now should succeed again
	if !ctrl.TryAcquire() {
		t.Error("expected acquire to succeed after release")
	}
}

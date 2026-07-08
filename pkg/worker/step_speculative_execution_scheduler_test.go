package worker

import (
	"testing"
	"time"
)

func TestSpeculativeExecutionScheduler(t *testing.T) {
	spec := SpeculativeExecutionSpec{
		P95Duration:     2 * time.Second,
		MinRunDuration:  3 * time.Second,
		SlowFactorMulti: 2.0, // threshold = 4s
	}
	scheduler := NewSpeculativeExecutionScheduler(spec)

	now := time.Now()
	// Started 2s ago (< min run 3s)
	if scheduler.ShouldSpeculate(now.Add(-2*time.Second), now) {
		t.Error("should not speculate below min run duration")
	}

	// Started 3.5s ago (< threshold 4s)
	if scheduler.ShouldSpeculate(now.Add(-3500*time.Millisecond), now) {
		t.Error("should not speculate below slow threshold")
	}

	// Started 4.5s ago (>= threshold 4s)
	if !scheduler.ShouldSpeculate(now.Add(-4500*time.Millisecond), now) {
		t.Error("should speculate straggler exceeding 4s")
	}
}

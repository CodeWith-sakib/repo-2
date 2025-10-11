package metrics

import (
	"testing"
	"time"
)

func TestSLATracker_BudgetAndBurnRate(t *testing.T) {
	// 99% target availability -> 1% error allowed
	cfg := SLATrackerConfig{
		TargetAvailability: 0.99,
		WindowDuration:     24 * time.Hour,
		ShortWindow:        time.Hour,
	}

	tracker := NewSLATracker(cfg)
	now := time.Now()

	// 98 successes, 2 failures -> total 100 requests (error rate = 2%)
	// Target error rate = 1%. Burn rate should be 2.0x!
	for i := 0; i < 98; i++ {
		tracker.RecordAt(now.Add(-time.Duration(i)*time.Second), true)
	}
	for i := 0; i < 2; i++ {
		tracker.RecordAt(now.Add(-time.Duration(i)*time.Second), false)
	}

	status := tracker.Status()

	if status.TotalRequests != 100 {
		t.Errorf("expected 100 requests, got %d", status.TotalRequests)
	}
	if status.FailedRequests != 2 {
		t.Errorf("expected 2 failed requests, got %d", status.FailedRequests)
	}
	if status.Availability != 0.98 {
		t.Errorf("expected availability 0.98, got %v", status.Availability)
	}
	// Burn rate: 2% observed / 1% allowed = 2.0x
	if status.BurnRateShort < 1.9 || status.BurnRateShort > 2.1 {
		t.Errorf("expected burn rate ~2.0, got %v", status.BurnRateShort)
	}
	if !status.IsExhausted {
		t.Error("expected error budget to be exhausted with 2% failures vs 1% allowance")
	}
	if status.String() == "" {
		t.Error("String() should not be empty")
	}
}

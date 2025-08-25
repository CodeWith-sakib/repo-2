package postgres

import (
	"testing"
	"time"
)

func TestPoolTelemetryProbe_SnapshotMetrics(t *testing.T) {
	probe := NewPoolTelemetryProbe(10, 30*time.Second)

	waits := []time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
		15 * time.Millisecond,
		20 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, w := range waits {
		probe.RecordAcquire(w)
	}
	probe.RecordRelease()
	probe.RecordFailure()

	snap := probe.Snapshot()

	if snap.ActiveConns != 4 {
		t.Errorf("expected 4 active conns, got %d", snap.ActiveConns)
	}
	if snap.AcquiredTotal != 5 {
		t.Errorf("expected 5 acquired, got %d", snap.AcquiredTotal)
	}
	if snap.FailedTotal != 1 {
		t.Errorf("expected 1 failure, got %d", snap.FailedTotal)
	}
	if snap.P50WaitMs <= 0 {
		t.Errorf("expected positive P50 wait, got %.2f", snap.P50WaitMs)
	}
	if snap.P99WaitMs < snap.P50WaitMs {
		t.Errorf("P99 should be >= P50: p99=%.2f p50=%.2f", snap.P99WaitMs, snap.P50WaitMs)
	}
	if snap.SaturationRatio < 0 || snap.SaturationRatio > 1.0 {
		t.Errorf("saturation ratio out of range: %v", snap.SaturationRatio)
	}
	if len(snap.String()) == 0 {
		t.Error("snapshot String() must not be empty")
	}
}

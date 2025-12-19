package worker

import (
	"testing"
	"time"
)

func TestWorkerWatchdog(t *testing.T) {
	wd := NewWorkerWatchdog(5 * time.Second)

	wd.RecordHeartbeat("w1", 2)
	wd.RecordHeartbeat("w2", 0)

	if !wd.IsAvailable("w1") || !wd.IsAvailable("w2") {
		t.Fatal("expected both workers to be available initially")
	}

	// Fast-forward time past TTL
	simulatedNow := time.Now().UTC().Add(10 * time.Second)
	cordoned := wd.Sweep(simulatedNow)

	if len(cordoned) != 2 {
		t.Fatalf("expected 2 cordoned workers, got %d", len(cordoned))
	}

	if wd.IsAvailable("w1") {
		t.Error("expected w1 to be isolated")
	}

	// Test auto-recovery on heartbeat
	wd.RecordHeartbeat("w1", 1)
	if !wd.IsAvailable("w1") {
		t.Error("expected w1 to be available after heartbeat recovery")
	}
}

package postgres

import (
	"testing"
	"time"
)

func TestConnectionHealthReaper_LeakedDetection(t *testing.T) {
	cfg := ConnectionHealthConfig{
		MaxLeaseDuration:    20 * time.Millisecond,
		MaxIdleDuration:     time.Hour,
		MaxConsecutiveError: 2,
	}

	reaper := NewConnectionHealthReaper(cfg)
	reaper.RegisterConnection(101)
	_ = reaper.RecordAcquire(101, "worker-job-a")

	// Wait for lease to expire
	time.Sleep(30 * time.Millisecond)

	leaked, _ := reaper.ReapStaleConnections()
	if len(leaked) != 1 || leaked[0] != 101 {
		t.Fatalf("expected connection 101 to be reaped as leaked, got %v", leaked)
	}

	if reaper.ReapedLeakCount() != 1 {
		t.Errorf("expected 1 reaped leak, got %d", reaper.ReapedLeakCount())
	}
}

func TestConnectionHealthReaper_ConsecutiveErrors(t *testing.T) {
	cfg := ConnectionHealthConfig{
		MaxLeaseDuration:    time.Hour,
		MaxConsecutiveError: 3,
	}

	reaper := NewConnectionHealthReaper(cfg)
	reaper.RegisterConnection(202)

	reaper.RecordError(202)
	reaper.RecordError(202)

	_, dead := reaper.ReapStaleConnections()
	if len(dead) != 0 {
		t.Errorf("connection should not be dead after 2 errors (threshold is 3)")
	}

	reaper.RecordError(202) // 3rd error -> triggers suspect -> dead
	_, dead = reaper.ReapStaleConnections()
	if len(dead) != 1 || dead[0] != 202 {
		t.Fatalf("expected connection 202 to be dead, got %v", dead)
	}
}

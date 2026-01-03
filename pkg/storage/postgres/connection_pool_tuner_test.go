package postgres

import (
	"testing"
	"time"
)

func TestConnectionPoolTuner(t *testing.T) {
	tuner := NewConnectionPoolTuner()

	prof := tuner.ComputeProfile(8, 50, true)
	if prof.MaxOpenConns <= 0 || prof.MaxIdleConns <= 0 {
		t.Fatalf("expected positive pool sizes, got open=%d idle=%d", prof.MaxOpenConns, prof.MaxIdleConns)
	}
	if prof.MaxIdleConns > prof.MaxOpenConns {
		t.Errorf("max idle %d cannot exceed max open %d", prof.MaxIdleConns, prof.MaxOpenConns)
	}
	if prof.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("expected 30m max lifetime, got %v", prof.ConnMaxLifetime)
	}

	// Test nil db apply without panic
	tuner.ApplyConfig(nil, prof)
}

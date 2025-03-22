package retry

import (
	"testing"
	"time"
)

func TestSimulateBackoff(t *testing.T) {
	pol := Policy{
		MaxAttempts:     5,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		BackoffFactor:   2.0,
		Jitter:          false,
	}

	res := SimulateBackoff(pol, 4)
	if res.Attempts != 4 {
		t.Errorf("expected 4 attempts, got %d", res.Attempts)
	}
	if res.MaxInterval < 10*time.Millisecond {
		t.Errorf("unexpected max interval: %v", res.MaxInterval)
	}
	sd := StandardDeviation(res)
	if sd < 0 {
		t.Errorf("invalid standard deviation: %v", sd)
	}
}

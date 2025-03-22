package retry

import (
	"testing"
	"time"
)

func TestSimulateBackoff(t *testing.T) {
	strategy := NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond, 2.0, false)

	res := SimulateBackoff(strategy, 4)
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

package worker

import (
	"errors"
	"testing"
)

func TestCPUAffinityGroup(t *testing.T) {
	group := NewCPUAffinityGroup(4)

	cores1, err := group.AllocateCores("step-1", 2)
	if err != nil {
		t.Fatalf("unexpected allocation error: %v", err)
	}
	if len(cores1) != 2 {
		t.Errorf("expected 2 cores, got %d", len(cores1))
	}
	if group.Utilization() != 0.5 {
		t.Errorf("expected 50 percent utilization, got %f", group.Utilization())
	}

	// Try allocating 3 more cores (only 2 left)
	_, err = group.AllocateCores("step-2", 3)
	if !errors.Is(err, ErrNoAvailableCPUCores) {
		t.Fatalf("expected ErrNoAvailableCPUCores, got %v", err)
	}

	group.ReleaseCores("step-1")
	if group.Utilization() != 0.0 {
		t.Errorf("expected 0 utilization after release, got %f", group.Utilization())
	}
}

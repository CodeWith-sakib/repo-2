package worker

import (
	"context"
	"testing"
	"time"
)

func TestDrainCoordinatorLifecycle(t *testing.T) {
	dc := NewDrainCoordinator()

	flushed := false
	dc.RegisterFlushHook(func(ctx context.Context) error {
		flushed = true
		return nil
	})

	if !dc.IncActive() {
		t.Fatal("expected IncActive to succeed in normal phase")
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		dc.DecActive()
	}()

	ctx := context.Background()
	if err := dc.StartDrain(ctx, 1*time.Second); err != nil {
		t.Fatalf("drain failed: %v", err)
	}

	if dc.Phase() != DrainPhaseCompleted {
		t.Errorf("expected COMPLETED phase, got %s", dc.Phase())
	}
	if !flushed {
		t.Error("expected flush hook to run")
	}
	if dc.IncActive() {
		t.Error("expected IncActive to be rejected after drain completed")
	}
}

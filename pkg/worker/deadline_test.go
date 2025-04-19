package worker

import (
	"context"
	"testing"
	"time"
)

func TestDeadlineCoordinator(t *testing.T) {
	coord := NewDeadlineCoordinator()
	budget := ExecutionBudget{
		StepTimeout: 50 * time.Millisecond,
	}

	ctx, cancel := coord.WrapContext(context.Background(), "task-1", budget)
	defer cancel()

	rem, exists := coord.RemainingTime("task-1")
	if !exists || rem <= 0 {
		t.Errorf("expected positive remaining time, got %v", rem)
	}

	select {
	case <-ctx.Done():
		t.Error("context expired prematurely")
	case <-time.After(10 * time.Millisecond):
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("expected context to be cancelled after deadline")
	}
}

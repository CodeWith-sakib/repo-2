package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkflowTimeBudgetController(t *testing.T) {
	ctrl := NewWorkflowTimeBudgetController("wf-budget-1", 1*time.Second)

	ctrl.SetStepLimit("quick-step", 200*time.Millisecond)

	ctx, cancel, err := ctrl.AllocateStepContext(context.Background(), "quick-step")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}

	maxExpected := time.Now().Add(250 * time.Millisecond)
	if deadline.After(maxExpected) {
		t.Errorf("expected deadline within 250ms, got %v", deadline)
	}

	ctrl.RecordStepElapsed("quick-step", 150*time.Millisecond)
	if ctrl.RemainingBudget() <= 0 {
		t.Error("expected remaining budget > 0")
	}

	// Exhaust budget
	ctrl.startTime = time.Now().Add(-2 * time.Second)
	_, _, err = ctrl.AllocateStepContext(context.Background(), "late-step")
	if !errors.Is(err, ErrTimeBudgetExhausted) {
		t.Errorf("expected ErrTimeBudgetExhausted, got %v", err)
	}
}

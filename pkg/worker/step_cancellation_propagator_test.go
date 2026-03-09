package worker

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestStepCancellationPropagator(t *testing.T) {
	prop := NewStepCancellationPropagator()

	var pStepCancelled int64
	pCtx, pCancel := context.WithCancel(context.Background())
	_ = pCtx
	prop.RegisterStep("run-parent", "step-1", func() {
		atomic.AddInt64(&pStepCancelled, 1)
		pCancel()
	})

	var cStepCancelled int64
	cCtx, cCancel := context.WithCancel(context.Background())
	_ = cCtx
	prop.RegisterStep("run-child", "step-a", func() {
		atomic.AddInt64(&cStepCancelled, 1)
		cCancel()
	})

	prop.LinkChildRun("run-parent", "run-child")

	// Cascading cancel on run-parent should cancel both parent and child steps
	count := prop.CancelCascade("run-parent")
	if count != 2 {
		t.Errorf("expected 2 cancelled steps, got %d", count)
	}

	if atomic.LoadInt64(&pStepCancelled) != 1 || atomic.LoadInt64(&cStepCancelled) != 1 {
		t.Errorf("expected both cancellation callbacks invoked")
	}
}

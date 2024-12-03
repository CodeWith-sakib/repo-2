package statemachine

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestSagaCompensationReverseOrder(t *testing.T) {
	sm := NewSagaManager()
	runID := core.NewID("saga-run")

	sm.RegisterCompensation(runID, &CompensationAction{StepID: "step-1"})
	sm.RegisterCompensation(runID, &CompensationAction{StepID: "step-2"})
	sm.RegisterCompensation(runID, &CompensationAction{StepID: "step-3"})

	var executed []string
	err := sm.Compensate(context.Background(), runID, func(ctx context.Context, action *CompensationAction) error {
		executed = append(executed, action.StepID)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected compensation error: %v", err)
	}

	expected := []string{"step-3", "step-2", "step-1"}
	if len(executed) != len(expected) {
		t.Fatalf("length mismatch: got %d, expected %d", len(executed), len(expected))
	}
	for i, v := range expected {
		if executed[i] != v {
			t.Errorf("at index %d: got %s, expected %s", i, executed[i], v)
		}
	}
}

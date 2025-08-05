package statemachine

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestStateTransitionJournal(t *testing.T) {
	journal := NewStateTransitionJournal()
	runID := core.NewID("run-1")

	journal.RecordTransition(runID, core.RunStatePending, core.RunStateRunning, "scheduled by worker")
	journal.RecordTransition(runID, core.RunStateRunning, core.RunStateCompleted, "all steps finished")

	hist := journal.HistoryForRun(runID)
	if len(hist) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(hist))
	}
	if hist[0].FromState != core.RunStatePending || hist[1].ToState != core.RunStateCompleted {
		t.Errorf("unexpected transition progression: %+v", hist)
	}
}

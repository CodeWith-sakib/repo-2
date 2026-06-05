package worker

import (
	"testing"
)

func TestCancellationCascadeEvaluator(t *testing.T) {
	evaluator := NewCancellationCascadeEvaluator()

	adj := map[string][]string{
		"step-1": {"step-2", "step-3"},
		"step-2": {"step-4"},
		"step-3": {"step-4"},
		"step-4": {"step-5"},
		"step-5": {},
	}

	affected := evaluator.ComputeAffectedSteps("step-2", adj)
	expectedMap := map[string]bool{"step-4": true, "step-5": true}

	if len(affected) != len(expectedMap) {
		t.Fatalf("expected %d affected steps, got %d (%v)", len(expectedMap), len(affected), affected)
	}
	for _, id := range affected {
		if !expectedMap[id] {
			t.Errorf("unexpected step in cascade: %s", id)
		}
	}
}

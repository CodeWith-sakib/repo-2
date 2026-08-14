package core

import (
	"context"
	"testing"
)

func TestDependencyGraphPruner(t *testing.T) {
	pruner := NewDependencyGraphPruner()

	adj := map[string][]string{
		"step-1": {"step-2"},
		"step-2": {"step-3"},
		"step-3": {},
	}
	enabled := map[string]bool{
		"step-1": true,
		"step-2": false, // disabled intermediate node
		"step-3": true,
	}

	plan := pruner.PruneDisabledSteps(context.Background(), adj, enabled)

	if len(plan.RetainedNodes) != 2 {
		t.Errorf("expected 2 retained nodes, got %d", len(plan.RetainedNodes))
	}
	if len(plan.PrunedNodes) != 1 || plan.PrunedNodes[0] != "step-2" {
		t.Errorf("expected step-2 pruned, got %v", plan.PrunedNodes)
	}

	// Step-1 should now directly link to step-3
	targets := plan.CompactedEdges["step-1"]
	if len(targets) != 1 || targets[0] != "step-3" {
		t.Errorf("expected compacted edge step-1 -> step-3, got %v", targets)
	}
}

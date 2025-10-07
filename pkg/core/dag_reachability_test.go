package core

import (
	"testing"
)

func TestDAGReachabilityIndex_Basic(t *testing.T) {
	// A -> B -> C -> D
	// E -> C
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell"},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "E", TaskType: "shell"},
		{ID: "C", TaskType: "shell", DependsOn: []string{"B", "E"}},
		{ID: "D", TaskType: "shell", DependsOn: []string{"C"}},
	}

	idx, err := BuildReachabilityIndex(steps)
	if err != nil {
		t.Fatalf("build index failed: %v", err)
	}

	// A can reach B, C, D
	if !idx.CanReach("A", "B") {
		t.Error("expected A can reach B")
	}
	if !idx.CanReach("A", "C") {
		t.Error("expected A can reach C (transitive)")
	}
	if !idx.CanReach("A", "D") {
		t.Error("expected A can reach D (transitive)")
	}

	// E can reach C, D
	if !idx.CanReach("E", "C") {
		t.Error("expected E can reach C")
	}
	if !idx.CanReach("E", "D") {
		t.Error("expected E can reach D")
	}

	// E cannot reach A or B
	if idx.CanReach("E", "A") || idx.CanReach("E", "B") {
		t.Error("E should not reach A or B")
	}

	// D cannot reach anything (terminal)
	if idx.CanReach("D", "A") || idx.CanReach("D", "C") {
		t.Error("D should not reach any predecessors")
	}

	// Ancestor count for D should be 4 (A, B, E, C)
	if idx.AncestorCount("D") != 4 {
		t.Errorf("expected 4 ancestors for D, got %d", idx.AncestorCount("D"))
	}
}

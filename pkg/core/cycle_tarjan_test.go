package core

import (
	"testing"
)

func TestTarjanSCCDetector(t *testing.T) {
	// A -> B -> C -> A (cycle of 3 nodes)
	steps := []StepDefinition{
		{ID: "A", DependsOn: []string{"C"}},
		{ID: "B", DependsOn: []string{"A"}},
		{ID: "C", DependsOn: []string{"B"}},
		{ID: "D", DependsOn: []string{"C"}}, // Downstream outside cycle
	}

	dag, _ := BuildDAG(steps)
	detector := NewTarjanSCCDetector(dag)
	cycles := detector.FindCycles()

	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle detected, got %d", len(cycles))
	}

	if len(cycles[0]) != 3 {
		t.Errorf("expected 3 nodes in detected cycle, got %v", cycles[0])
	}
}

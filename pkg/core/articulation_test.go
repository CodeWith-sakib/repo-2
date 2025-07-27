package core

import (
	"testing"
)

func TestArticulationPointDetector(t *testing.T) {
	// Graph: A -> B -> C, where B is a bridge bottleneck (articulation point)
	steps := []StepDefinition{
		{ID: "A"},
		{ID: "B", DependsOn: []string{"A"}},
		{ID: "C", DependsOn: []string{"B"}},
	}

	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	detector := NewArticulationPointDetector(dag)
	cuts := detector.FindCutVertices()

	if len(cuts) != 1 || cuts[0] != "B" {
		t.Errorf("expected cut vertex [B], got %v", cuts)
	}
}

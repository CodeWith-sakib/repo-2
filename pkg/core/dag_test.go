package core

import (
	"testing"
)

func TestDAGTopologicalSort(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-c", TaskType: "shell", DependsOn: []string{"step-a", "step-b"}},
		{ID: "step-a", TaskType: "http"},
		{ID: "step-b", TaskType: "shell", DependsOn: []string{"step-a"}},
		{ID: "step-d", TaskType: "http", DependsOn: []string{"step-c"}},
	}

	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("unexpected error building DAG: %v", err)
	}

	if err := dag.ValidateAcyclic(); err != nil {
		t.Fatalf("expected acyclic, got: %v", err)
	}

	order, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}

	expected := []string{"step-a", "step-b", "step-c", "step-d"}
	if len(order) != len(expected) {
		t.Fatalf("length mismatch: got %d, expected %d", len(order), len(expected))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: got %s, expected %s", i, order[i], v)
		}
	}
}

func TestDAGCycleDetection(t *testing.T) {
	cyclicSteps := []StepDefinition{
		{ID: "node-1", TaskType: "shell", DependsOn: []string{"node-3"}},
		{ID: "node-2", TaskType: "shell", DependsOn: []string{"node-1"}},
		{ID: "node-3", TaskType: "shell", DependsOn: []string{"node-2"}},
	}

	dag, err := BuildDAG(cyclicSteps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := dag.ValidateAcyclic(); err == nil {
		t.Fatal("expected cycle error, got nil")
	}

	if _, err := dag.TopologicalSort(); err == nil {
		t.Fatal("expected topological sort to fail on cycle")
	}
}

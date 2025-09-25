package core

import (
	"reflect"
	"testing"
)

func TestDAGOptimizer_TransitiveReduction(t *testing.T) {
	// Graph:
	// A -> B -> C
	// A -> C (redundant edge!)
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell"},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "C", TaskType: "shell", DependsOn: []string{"A", "B"}},
	}

	opt := NewDAGOptimizer()
	optimized, report, err := opt.Optimize(steps)
	if err != nil {
		t.Fatalf("optimize failed: %v", err)
	}

	if report.RemovedEdges != 1 {
		t.Errorf("expected 1 removed edge, got %d", report.RemovedEdges)
	}

	// C should now only depend on B
	for _, s := range optimized {
		if s.ID == "C" {
			expectedDeps := []string{"B"}
			if !reflect.DeepEqual(s.DependsOn, expectedDeps) {
				t.Errorf("expected C deps %v, got %v", expectedDeps, s.DependsOn)
			}
		}
	}
}

func TestDAGOptimizer_RootsAndTerminals(t *testing.T) {
	steps := []StepDefinition{
		{ID: "root1", TaskType: "shell"},
		{ID: "root2", TaskType: "shell"},
		{ID: "mid", TaskType: "shell", DependsOn: []string{"root1", "root2"}},
		{ID: "leaf1", TaskType: "shell", DependsOn: []string{"mid"}},
		{ID: "leaf2", TaskType: "shell", DependsOn: []string{"mid"}},
	}

	opt := NewDAGOptimizer()
	roots := opt.FindRoots(steps)
	terminals := opt.FindTerminals(steps)

	expectedRoots := []string{"root1", "root2"}
	if !reflect.DeepEqual(roots, expectedRoots) {
		t.Errorf("expected roots %v, got %v", expectedRoots, roots)
	}

	expectedTerminals := []string{"leaf1", "leaf2"}
	if !reflect.DeepEqual(terminals, expectedTerminals) {
		t.Errorf("expected terminals %v, got %v", expectedTerminals, terminals)
	}
}

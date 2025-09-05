package core

import "fmt"

import (
	"testing"
)

func TestPartitionDAG_LinearChain(t *testing.T) {
	// A → B → C: should produce 3 levels, each with 1 step
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell"},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "C", TaskType: "shell", DependsOn: []string{"B"}},
	}

	parts, err := PartitionDAG(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 3 {
		t.Errorf("expected 3 levels, got %d", len(parts))
	}
	for _, p := range parts {
		if len(p.Steps) != 1 {
			t.Errorf("expected 1 step per level, got %d at level %d", len(p.Steps), p.Level)
		}
	}
	bns := Bottlenecks(parts)
	if len(bns) != 3 {
		t.Errorf("expected 3 bottlenecks (all serial), got %d", len(bns))
	}
}

func TestPartitionDAG_DiamondShape(t *testing.T) {
	// A → {B, C} → D: 3 levels, middle has 2 parallel steps
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell"},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "C", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "D", TaskType: "shell", DependsOn: []string{"B", "C"}},
	}

	parts, err := PartitionDAG(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 3 {
		t.Errorf("expected 3 levels, got %d", len(parts))
	}
	if len(parts[1].Steps) != 2 {
		t.Errorf("expected 2 parallel steps at level 1, got %d", len(parts[1].Steps))
	}

	pf := ParallelismFactor(parts)
	if pf < 1.0 {
		t.Errorf("parallelism factor should be > 1.0, got %.2f", pf)
	}
}

func TestPartitionDAG_CycleDetected(t *testing.T) {
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell", DependsOn: []string{"C"}},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "C", TaskType: "shell", DependsOn: []string{"B"}},
	}
	if _, err := PartitionDAG(steps); err == nil {
		t.Error("expected cycle error")
	}
}

func TestPartitionDAG_WideParallel(t *testing.T) {
	// 1 root → 10 parallel leaves
	steps := []StepDefinition{{ID: "root", TaskType: "shell"}}
	for i := 0; i < 10; i++ {
		steps = append(steps, StepDefinition{
			ID:        fmt.Sprintf("leaf-%d", i),
			TaskType:  "shell",
			DependsOn: []string{"root"},
		})
	}

	parts, err := PartitionDAG(steps)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(parts[1].Steps) != 10 {
		t.Errorf("expected 10 parallel leaves at level 1, got %d", len(parts[1].Steps))
	}
}

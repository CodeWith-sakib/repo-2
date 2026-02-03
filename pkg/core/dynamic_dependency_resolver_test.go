package core

import (
	"testing"
)

func TestDynamicDependencyResolver(t *testing.T) {
	steps := []StepDefinition{
		{ID: "A"},
		{ID: "B", DependsOn: []string{"A"}},
		{ID: "C", DependsOn: []string{"A"}},
		{ID: "D", DependsOn: []string{"B", "C"}},
	}

	resolver := NewDynamicDependencyResolver(steps)
	waves, err := resolver.ResolveExecutionWaves()
	if err != nil {
		t.Fatalf("unexpected resolution error: %v", err)
	}

	if len(waves) != 3 {
		t.Fatalf("expected 3 waves, got %d", len(waves))
	}

	if len(waves[0]) != 1 || waves[0][0] != "A" {
		t.Errorf("expected wave 0 to contain [A], got %v", waves[0])
	}
	if len(waves[1]) != 2 {
		t.Errorf("expected wave 1 to contain 2 steps (B, C), got %v", waves[1])
	}
	if len(waves[2]) != 1 || waves[2][0] != "D" {
		t.Errorf("expected wave 2 to contain [D], got %v", waves[2])
	}

	// Test cycle detection
	cycleSteps := []StepDefinition{
		{ID: "X", DependsOn: []string{"Z"}},
		{ID: "Y", DependsOn: []string{"X"}},
		{ID: "Z", DependsOn: []string{"Y"}},
	}
	cycleResolver := NewDynamicDependencyResolver(cycleSteps)
	_, err = cycleResolver.ResolveExecutionWaves()
	if err == nil {
		t.Error("expected error for cycle, got nil")
	}
}

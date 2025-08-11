package core

import (
	"testing"
	"time"
)

func TestCriticalPathEngine_SlackCalculation(t *testing.T) {
	// Linear DAG: A -> B -> C
	// Branch: A -> D -> C
	// A=2s, B=5s, C=1s (total 8s)
	// D=1s (A(2)+D(1)=3 -> finishes at 3s, but C waits till 7s -> D slack = 4s)
	steps := []StepDefinition{
		{ID: "A", TaskType: "shell"},
		{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
		{ID: "C", TaskType: "shell", DependsOn: []string{"B", "D"}},
		{ID: "D", TaskType: "shell", DependsOn: []string{"A"}},
	}

	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("failed to build dag: %v", err)
	}

	durations := map[string]time.Duration{
		"A": 2 * time.Second,
		"B": 5 * time.Second,
		"C": 1 * time.Second,
		"D": 1 * time.Second,
	}

	engine := NewCriticalPathEngine(dag)
	slackMap, err := engine.CalculateSlack(durations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slackMap["A"].IsCritical || !slackMap["B"].IsCritical || !slackMap["C"].IsCritical {
		t.Errorf("A, B, C must be on critical path")
	}

	if slackMap["D"].IsCritical {
		t.Errorf("D must not be critical")
	}

	if slackMap["D"].TotalSlack != 4*time.Second {
		t.Errorf("expected D slack to be 4s, got %v", slackMap["D"].TotalSlack)
	}
}

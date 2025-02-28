package core

import (
	"testing"
	"time"
)

func TestAnalyzeCriticalPath(t *testing.T) {
	steps := []StepDefinition{
		{ID: "A", DependsOn: []string{}},
		{ID: "B", DependsOn: []string{"A"}},
		{ID: "C", DependsOn: []string{"A"}},
		{ID: "D", DependsOn: []string{"B", "C"}},
	}
	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("failed building dag: %v", err)
	}

	durations := map[string]time.Duration{
		"A": 10 * time.Second,
		"B": 20 * time.Second,
		"C": 5 * time.Second,
		"D": 10 * time.Second,
	}

	cpa := AnalyzeCriticalPath(dag, durations)
	// Critical path: A -> B -> D (10 + 20 + 10 = 40s)
	if cpa.TotalDuration != 40*time.Second {
		t.Errorf("expected total duration 40s, got %v", cpa.TotalDuration)
	}

	if !cpa.Schedule["B"].IsCritical {
		t.Error("expected step B to be critical")
	}
	if cpa.Schedule["C"].IsCritical {
		t.Error("expected step C not to be critical")
	}
}

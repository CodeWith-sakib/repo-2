package core

import (
	"strings"
	"testing"
)

func TestAdvancedDAGValidator_Healthy(t *testing.T) {
	steps := []StepDefinition{
		{ID: "fetch", TaskType: "http"},
		{ID: "parse", TaskType: "json", DependsOn: []string{"fetch"}},
		{ID: "store", TaskType: "sql", DependsOn: []string{"parse"}},
	}

	v := NewAdvancedDAGValidator()
	report := v.ValidateGraph(steps)

	if !report.IsValid {
		t.Fatalf("expected healthy graph to be valid, got issues: %v", report.Issues)
	}
	if report.ConnectedComponents != 1 {
		t.Errorf("expected 1 connected component, got %d", report.ConnectedComponents)
	}
	if len(report.RootStepIDs) != 1 || report.RootStepIDs[0] != "fetch" {
		t.Errorf("expected root 'fetch', got %v", report.RootStepIDs)
	}
	if len(report.TerminalStepIDs) != 1 || report.TerminalStepIDs[0] != "store" {
		t.Errorf("expected terminal 'store', got %v", report.TerminalStepIDs)
	}
}

func TestAdvancedDAGValidator_DanglingAndDisconnected(t *testing.T) {
	steps := []StepDefinition{
		{ID: "stepA", TaskType: "shell", DependsOn: []string{"unknownStep"}}, // dangling
		{ID: "isolatedX", TaskType: "shell"},                                 // disconnected component
	}

	v := NewAdvancedDAGValidator()
	report := v.ValidateGraph(steps)

	if report.IsValid {
		t.Error("expected invalid report due to dangling reference")
	}

	hasDangling := false
	for _, is := range report.Issues {
		if is.Type == IssueDanglingReference {
			hasDangling = true
		}
	}
	if !hasDangling {
		t.Error("expected IssueDanglingReference in issues")
	}

	if report.ConnectedComponents != 2 {
		t.Errorf("expected 2 disconnected components, got %d", report.ConnectedComponents)
	}

	summary := report.Summary()
	if !strings.Contains(summary, "DAG Structural Health: INVALID") {
		t.Errorf("unexpected summary: %s", summary)
	}
}

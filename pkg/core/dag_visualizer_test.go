package core

import (
	"strings"
	"testing"
)

func TestGraphVisualizer_ToDOT(t *testing.T) {
	wf := &WorkflowDefinition{
		Name:    "PaymentFlow",
		Version: 1,
		Steps: []StepDefinition{
			{ID: "charge", TaskType: "http"},
			{ID: "receipt", TaskType: "email", DependsOn: []string{"charge"}},
		},
	}

	vis := NewGraphVisualizer()
	dot := vis.ToDOT(wf)

	if !strings.Contains(dot, `digraph "PaymentFlow"`) {
		t.Errorf("expected digraph header, got: %s", dot)
	}
	if !strings.Contains(dot, `"charge" -> "receipt"`) {
		t.Errorf("expected edge from charge to receipt, got: %s", dot)
	}
}

func TestGraphVisualizer_ToASCII(t *testing.T) {
	wf := &WorkflowDefinition{
		Name:    "LinearTree",
		Version: 2,
		Steps: []StepDefinition{
			{ID: "start", TaskType: "shell"},
			{ID: "middle", TaskType: "shell", DependsOn: []string{"start"}},
			{ID: "end", TaskType: "shell", DependsOn: []string{"middle"}},
		},
	}

	vis := NewGraphVisualizer()
	ascii := vis.ToASCII(wf)

	if !strings.Contains(ascii, "Workflow: LinearTree (v2)") {
		t.Errorf("missing header in ascii: %s", ascii)
	}
	if !strings.Contains(ascii, "• start") {
		t.Errorf("missing root in ascii: %s", ascii)
	}
	if !strings.Contains(ascii, "└─ middle") {
		t.Errorf("missing middle in ascii: %s", ascii)
	}
}

package core

import (
	"strings"
	"testing"
)

func TestDAGVisualizer(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-1", TaskType: "http"},
		{ID: "step-2", TaskType: "shell", DependsOn: []string{"step-1"}},
		{ID: "step-3", TaskType: "shell", DependsOn: []string{"step-2"}},
	}
	dag, _ := BuildDAG(steps)
	vis := NewDAGVisualizer(dag)

	dot := vis.ToDOT()
	if !strings.Contains(dot, "digraph WorkflowDAG") || !strings.Contains(dot, `"step-1" -> "step-2"`) {
		t.Errorf("unexpected DOT output: %s", dot)
	}

	mermaid := vis.ToMermaid()
	if !strings.Contains(mermaid, "graph LR") || !strings.Contains(mermaid, "step-1 --> step-2") {
		t.Errorf("unexpected Mermaid output: %s", mermaid)
	}

	ascii := vis.ToASCII()
	if !strings.Contains(ascii, "Workflow DAG:") || !strings.Contains(ascii, "[step-1 (http)]") {
		t.Errorf("unexpected ASCII output: %s", ascii)
	}
}

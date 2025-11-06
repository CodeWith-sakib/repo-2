package core

import (
	"strings"
	"testing"
)

func TestDynamicPipelineCompiler_DiamondCompilation(t *testing.T) {
	// A -> B, C -> D
	wf := &WorkflowDefinition{
		ID:   "wf-diamond",
		Name: "Diamond Workflow",
		Steps: []StepDefinition{
			{ID: "A", TaskType: "shell"},
			{ID: "B", TaskType: "shell", DependsOn: []string{"A"}},
			{ID: "C", TaskType: "shell", DependsOn: []string{"A"}},
			{ID: "D", TaskType: "shell", DependsOn: []string{"B", "C"}},
		},
	}

	compiler := NewDynamicPipelineCompiler()
	ir, err := compiler.Compile(wf)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	if ir.TotalStages != 3 {
		t.Fatalf("expected 3 stages (A -> [B,C] -> D), got %d", ir.TotalStages)
	}

	// Stage 0: A (Sequential)
	if ir.Stages[0].Parallel || len(ir.Stages[0].StepIDs) != 1 || ir.Stages[0].StepIDs[0] != "A" {
		t.Errorf("unexpected stage 0: %+v", ir.Stages[0])
	}

	// Stage 1: B, C (Parallel)
	if !ir.Stages[1].Parallel || len(ir.Stages[1].StepIDs) != 2 {
		t.Errorf("unexpected stage 1 (expected 2 parallel steps): %+v", ir.Stages[1])
	}

	// Stage 2: D (Sequential)
	if ir.Stages[2].Parallel || len(ir.Stages[2].StepIDs) != 1 || ir.Stages[2].StepIDs[0] != "D" {
		t.Errorf("unexpected stage 2: %+v", ir.Stages[2])
	}

	formatted := ir.FormatIR()
	if !strings.Contains(formatted, "Stage [1] (Parallel, 2 steps): B, C") {
		t.Errorf("unexpected formatted output: %s", formatted)
	}
}

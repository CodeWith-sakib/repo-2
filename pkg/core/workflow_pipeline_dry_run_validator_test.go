package core

import (
	"context"
	"errors"
	"testing"
)

func TestPipelineDryRunValidator(t *testing.T) {
	val := NewPipelineDryRunValidator([]string{"compute", "store"})

	validDef := &WorkflowDefinition{
		ID: "wf-val-1",
		Steps: []StepDefinition{
			{ID: "s1", TaskType: "compute"},
			{ID: "s2", TaskType: "store", DependsOn: []string{"s1"}},
		},
	}

	diags, err := val.ValidatePipeline(context.Background(), validDef)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}

	// Broken def (missing dep)
	brokenDef := &WorkflowDefinition{
		ID: "wf-val-2",
		Steps: []StepDefinition{
			{ID: "s1", TaskType: "compute", DependsOn: []string{"s_unknown"}},
		},
	}
	diags, err = val.ValidatePipeline(context.Background(), brokenDef)
	if !errors.Is(err, ErrValidationPipelineHalted) {
		t.Fatalf("expected ErrValidationPipelineHalted, got %v", err)
	}
	if len(diags) != 1 || diags[0].Severity != "ERROR" {
		t.Errorf("expected 1 error diagnostic, got %v", diags)
	}
}

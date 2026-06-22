package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSyntheticBiologyCRISPRCascadePipeline(t *testing.T) {
	dag := SyntheticBiologyCRISPRCascadePipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "synthetic-genomics-foundry" {
		t.Errorf("expected tenant synthetic-genomics-foundry, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &CrRNAProcessingHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "num_spacers") {
		t.Errorf("output missing num_spacers")
	}

	h3 := &CFDOffTargetScoreHandler{}
	out3, err := h3.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 3 failed: %v", err)
	}
	if !strings.Contains(string(out3.Output), "cleavage_specificity_passed") {
		t.Errorf("output missing cleavage_specificity_passed")
	}
}

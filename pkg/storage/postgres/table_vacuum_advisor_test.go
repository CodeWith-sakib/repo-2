package postgres

import (
	"context"
	"testing"
)

func TestTableVacuumAdvisor(t *testing.T) {
	advisor := NewTableVacuumAdvisor(nil)

	res, err := advisor.AnalyzeTableDeadTuples(context.Background(), 15.0)
	if err != nil {
		t.Fatalf("unexpected analysis error: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 tables analyzed, got %d", len(res))
	}

	if !res[0].RequiresVacuum {
		t.Error("expected workflow_run_steps to require vacuum")
	}

	if res[1].RequiresVacuum {
		t.Error("expected workflow_definitions to not require vacuum")
	}
}

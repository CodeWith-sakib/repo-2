package postgres

import (
	"context"
	"testing"
)

func TestRowVersionGCEstimator(t *testing.T) {
	estimator := NewRowVersionGCEstimator(0.25)

	s1 := estimator.RecordStats(context.Background(), "kf_workflows", 10000, 1000)
	if s1.NeedsVacuum {
		t.Error("10% dead tuples should not trigger vacuum")
	}

	s2 := estimator.RecordStats(context.Background(), "kf_step_runs", 10000, 4000)
	if !s2.NeedsVacuum {
		t.Error("28% dead tuples should trigger vacuum")
	}

	candidates := estimator.CandidateVacuumTables()
	if len(candidates) != 1 || candidates[0].TableName != "kf_step_runs" {
		t.Errorf("expected kf_step_runs in candidates, got %v", candidates)
	}
}

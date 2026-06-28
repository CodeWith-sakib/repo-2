package postgres

import (
	"testing"
)

func TestIndexBloatEstimator(t *testing.T) {
	estimator := NewIndexBloatEstimator(0.30)

	// Healthy index
	rep1 := estimator.RecordIndexStats("kf_workflows", "idx_wf_tenant", 20*1024*1024, 18*1024*1024)
	if rep1.NeedsReindex {
		t.Error("10% bloat should not need reindex")
	}

	// Bloated index (> 30% bloat and > 10MB)
	rep2 := estimator.RecordIndexStats("kf_step_runs", "idx_steps_status", 50*1024*1024, 25*1024*1024)
	if !rep2.NeedsReindex {
		t.Error("50% bloat on 50MB index should need reindex")
	}

	candidates := estimator.CandidateReindexes()
	if len(candidates) != 1 || candidates[0].IndexName != "idx_steps_status" {
		t.Errorf("expected idx_steps_status in candidates, got %v", candidates)
	}
}

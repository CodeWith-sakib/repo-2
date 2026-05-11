package postgres

import (
	"testing"
	"time"
)

func TestQueryPlanCache(t *testing.T) {
	cache := NewQueryPlanCache(50 * time.Millisecond)

	plan := CachedPlanInfo{
		QueryFingerprint: "fp_workflow_runs_idx",
		EstimatedCost:    42.5,
		EstimatedRows:    1200,
		PlanTreeSummary:  "Index Scan using idx_runs on workflow_runs",
	}

	cache.StorePlan(plan)
	now := time.Now().UTC()

	retrieved, ok := cache.GetPlan("fp_workflow_runs_idx", now)
	if !ok || retrieved.EstimatedCost != 42.5 {
		t.Fatalf("expected to retrieve valid plan, got ok=%v", ok)
	}

	// Past TTL -> expired
	future := now.Add(60 * time.Millisecond)
	_, ok = cache.GetPlan("fp_workflow_runs_idx", future)
	if ok {
		t.Error("expected plan to be expired past TTL")
	}
}

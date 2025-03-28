package dashboard

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestBuildDashboardViewModel(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(-10 * time.Second)
	end := now

	runs := []*core.WorkflowRun{
		{
			ID:         core.NewID("run-1"),
			WorkflowID: core.NewID("wf-1"),
			State:      core.RunStateCompleted,
			StartedAt:  &start,
			FinishedAt: &end,
		},
		{
			ID:         core.NewID("run-2"),
			WorkflowID: core.NewID("wf-1"),
			State:      core.RunStateFailed,
			StartedAt:  &start,
			FinishedAt: &end,
		},
	}

	vm := BuildDashboardViewModel(runs)
	if vm.TotalRuns != 2 || vm.CompletedRuns != 1 || vm.FailedRuns != 1 {
		t.Errorf("unexpected counts: %+v", vm)
	}
	if vm.RecentRuns[0].BadgeClass != "badge-success" {
		t.Errorf("expected badge-success, got %s", vm.RecentRuns[0].BadgeClass)
	}
	if vm.RecentRuns[1].BadgeClass != "badge-danger" {
		t.Errorf("expected badge-danger, got %s", vm.RecentRuns[1].BadgeClass)
	}
}

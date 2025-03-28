package dashboard

import (
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type RunSummaryViewModel struct {
	ID         string
	WorkflowID string
	State      string
	Duration   string
	IsActive   bool
	BadgeClass string
}

type DashboardViewModel struct {
	TotalRuns     int
	ActiveRuns    int
	CompletedRuns int
	FailedRuns    int
	RecentRuns    []RunSummaryViewModel
}

func BuildDashboardViewModel(runs []*core.WorkflowRun) DashboardViewModel {
	vm := DashboardViewModel{
		TotalRuns:  len(runs),
		RecentRuns: make([]RunSummaryViewModel, 0, len(runs)),
	}

	for _, r := range runs {
		if r.State.IsActive() {
			vm.ActiveRuns++
		}
		if r.State == core.RunStateCompleted {
			vm.CompletedRuns++
		}
		if r.State == core.RunStateFailed {
			vm.FailedRuns++
		}

		durationStr := "-"
		if r.StartedAt != nil {
			if r.FinishedAt != nil {
				durationStr = r.FinishedAt.Sub(*r.StartedAt).Round(time.Millisecond).String()
			} else {
				durationStr = time.Since(*r.StartedAt).Round(time.Second).String()
			}
		}

		badge := "badge-info"
		switch r.State {
		case core.RunStateCompleted:
			badge = "badge-success"
		case core.RunStateFailed:
			badge = "badge-danger"
		case core.RunStateRunning:
			badge = "badge-primary"
		case core.RunStateCancelled:
			badge = "badge-secondary"
		}

		vm.RecentRuns = append(vm.RecentRuns, RunSummaryViewModel{
			ID:         string(r.ID),
			WorkflowID: string(r.WorkflowID),
			State:      string(r.State),
			Duration:   durationStr,
			IsActive:   r.State.IsActive(),
			BadgeClass: badge,
		})
	}

	return vm
}

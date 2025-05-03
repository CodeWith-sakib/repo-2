package dashboard

import (
	"strings"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestTextSummaryFormatter(t *testing.T) {
	fmttr := NewTextSummaryFormatter()

	now := time.Now().UTC()
	start := now.Add(-5 * time.Second)
	end := now

	run := &core.WorkflowRun{
		ID:              core.NewID("run"),
		WorkflowID:      core.NewID("wf"),
		WorkflowVersion: 1,
		State:           core.RunStateCompleted,
		StartedAt:       &start,
		FinishedAt:      &end,
	}

	steps := []*core.StepRun{
		{
			StepID:     "step-fetch",
			State:      core.StepStateCompleted,
			RetryCount: 1,
			StartedAt:  &start,
			FinishedAt: &end,
		},
	}

	md := fmttr.FormatMarkdown(run, steps)
	if !strings.Contains(md, "## Workflow Run:") || !strings.Contains(md, "| `step-fetch` | `COMPLETED` | 1 |") {
		t.Errorf("unexpected markdown output: %s", md)
	}
}

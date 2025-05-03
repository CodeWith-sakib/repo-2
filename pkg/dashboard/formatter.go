package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type TextSummaryFormatter struct{}

func NewTextSummaryFormatter() *TextSummaryFormatter {
	return &TextSummaryFormatter{}
}

func (f *TextSummaryFormatter) FormatMarkdown(run *core.WorkflowRun, steps []*core.StepRun) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Workflow Run: %s
", run.ID))
	sb.WriteString(fmt.Sprintf("- **State**: `%s`
", run.State))
	sb.WriteString(fmt.Sprintf("- **Workflow**: `%s` (v%d)
", run.WorkflowID, run.WorkflowVersion))

	duration := "N/A"
	if run.StartedAt != nil {
		if run.FinishedAt != nil {
			duration = run.FinishedAt.Sub(*run.StartedAt).Round(time.Millisecond).String()
		} else {
			duration = time.Since(*run.StartedAt).Round(time.Second).String()
		}
	}
	sb.WriteString(fmt.Sprintf("- **Duration**: %s

", duration))

	sb.WriteString("### Step Execution Breakdown
")
	sb.WriteString("| Step ID | State | Retries | Duration |
")
	sb.WriteString("| :--- | :--- | :--- | :--- |
")

	for _, s := range steps {
		stepDuration := "-"
		if s.StartedAt != nil && s.FinishedAt != nil {
			stepDuration = s.FinishedAt.Sub(*s.StartedAt).Round(time.Millisecond).String()
		}
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %d | %s |
", s.StepID, s.State, s.RetryCount, stepDuration))
	}

	return sb.String()
}

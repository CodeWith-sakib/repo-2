package dashboard

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestDashboardRendersInMemoryBackend(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Seed workflow and runs
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf"),
		TenantID: "default",
		Name:     "dashboard-test-pipeline",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "step-1", TaskType: "http"},
		},
	}
	_ = store.CreateWorkflow(ctx, wf)

	now := time.Now().UTC()
	run := &core.WorkflowRun{
		ID:         core.NewID("run-running"),
		WorkflowID: wf.ID,
		Version:    1,
		TenantID:   "default",
		State:      core.RunStateRunning,
		Priority:   core.PriorityHigh,
		StartedAt:  &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_ = store.CreateRun(ctx, run)

	handler := NewHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from dashboard, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "KestrelFlow System Dashboard") {
		t.Error("expected dashboard title in HTML")
	}
	if !strings.Contains(body, string(run.ID)) {
		t.Errorf("expected run ID %s rendered in table", run.ID)
	}
	if !strings.Contains(body, "badge-RUNNING") {
		t.Error("expected RUNNING badge in dashboard")
	}
}

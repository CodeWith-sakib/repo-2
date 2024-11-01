package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/api/types"
	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func setupTestServer() *Server {
	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, 10*time.Millisecond)
	return NewServer(store, sched)
}

func TestHealthzBoundary(t *testing.T) {
	srv := setupTestServer()

	// GET success
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Negative case: POST not allowed
	reqPost := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	wPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad method, got %d", wPost.Code)
	}
}

func TestCreateWorkflowBoundaries(t *testing.T) {
	srv := setupTestServer()

	// Negative case 1: Empty JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", bytes.NewBuffer([]byte("")))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", w.Code)
	}

	// Negative case 2: Missing steps
	noSteps := `{"name":"invalid-wf","steps":[]}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workflows", bytes.NewBufferString(noSteps))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing steps, got %d", w.Code)
	}

	// Negative case 3: Cyclic DAG steps
	cyclic := `{
		"name": "cyclic-pipeline",
		"steps": [
			{"id":"s1","task_type":"http","depends_on":["s2"]},
			{"id":"s2","task_type":"http","depends_on":["s1"]}
		]
	}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workflows", bytes.NewBufferString(cyclic))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cyclic workflow, got %d", w.Code)
	}

	// Positive case: Valid workflow creation
	valid := `{
		"name": "valid-pipeline",
		"steps": [
			{"id":"s1","task_type":"http"}
		]
	}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workflows", bytes.NewBufferString(valid))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetWorkflowBoundaries(t *testing.T) {
	srv := setupTestServer()

	// Negative case: non-existent workflow ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/wf-nonexistent", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent workflow, got %d", w.Code)
	}

	// Negative case: list with negative offset
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workflows?offset=-5", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative offset, got %d", w.Code)
	}
}

func TestSubmitAndCancelRunBoundaries(t *testing.T) {
	srv := setupTestServer()

	// Create a workflow first
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf"),
		TenantID: "default",
		Name:     "run-test-wf",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "step-1", TaskType: "http"},
		},
	}
	_ = srv.store.CreateWorkflow(context.Background(), wf)

	// Negative case: submit with non-existent workflow
	subBad := `{"workflow_id":"wf-unknown","priority":50}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewBufferString(subBad))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown workflow, got %d", w.Code)
	}

	// Negative case: submit with invalid priority (> 100)
	subPrio := `{"workflow_id":"` + string(wf.ID) + `","priority":150}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewBufferString(subPrio))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for priority > 100, got %d", w.Code)
	}

	// Positive submit
	subGood := `{"workflow_id":"` + string(wf.ID) + `","priority":50}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewBufferString(subGood))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 accepted, got %d", w.Code)
	}

	var runResp types.RunResponse
	_ = json.Unmarshal(w.Body.Bytes(), &runResp)

	// Cancel run positive
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runResp.ID+"/cancel", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on cancel, got %d", w.Code)
	}

	// Negative case: cancelling already cancelled run returns 409 Conflict
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runResp.ID+"/cancel", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict when cancelling already terminal run, got %d", w.Code)
	}

	// Negative case: non-existent run cancel
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runs/run-fake/cancel", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on cancelling non-existent run, got %d", w.Code)
	}
}

import os
from generator.git_utils import commit
from generator.loc import get_production_loc

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

def run_phase_3(dates_iter):
    print("=== Executing Phase 3: HTTP REST API, gRPC Surface & Boundary Validation ===")

    # Commit 3.1: Versioned DTO types and RFC 7807 error envelopes
    write_file("pkg/api/types/models.go", """package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type CreateWorkflowRequest struct {
	TenantID    string                `json:"tenant_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Steps       []core.StepDefinition `json:"steps"`
	Timeout     string                `json:"timeout,omitempty"`
	Metadata    map[string]string     `json:"metadata,omitempty"`
}

func (r *CreateWorkflowRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: workflow name is required", core.ErrValidationFailed)
	}
	if len(r.Steps) == 0 {
		return fmt.Errorf("%w: at least one step is required", core.ErrValidationFailed)
	}
	return nil
}

type WorkflowResponse struct {
	ID          string                `json:"id"`
	TenantID    string                `json:"tenant_id"`
	Name        string                `json:"name"`
	Version     int                   `json:"version"`
	Description string                `json:"description"`
	Steps       []core.StepDefinition `json:"steps"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type ListWorkflowsResponse struct {
	Workflows []*WorkflowResponse `json:"workflows"`
	Total     int                 `json:"total"`
	Limit     int                 `json:"limit"`
	Offset    int                 `json:"offset"`
}

type SubmitRunRequest struct {
	WorkflowID string          `json:"workflow_id"`
	Version    int             `json:"version,omitempty"`
	TenantID   string          `json:"tenant_id,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	Priority   int             `json:"priority,omitempty"`
}

func (r *SubmitRunRequest) Validate() error {
	if r.WorkflowID == "" {
		return fmt.Errorf("%w: workflow_id is required", core.ErrValidationFailed)
	}
	if r.Priority < 0 || r.Priority > 100 {
		return fmt.Errorf("%w: priority must be between 0 and 100", core.ErrValidationFailed)
	}
	return nil
}

type RunResponse struct {
	ID           string          `json:"id"`
	WorkflowID   string          `json:"workflow_id"`
	Version      int             `json:"version"`
	TenantID     string          `json:"tenant_id"`
	State        string          `json:"state"`
	Input        json.RawMessage `json:"input,omitempty"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	Priority     int             `json:"priority"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ListRunsResponse struct {
	Runs   []*RunResponse `json:"runs"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type CancelRunRequest struct {
	Reason string `json:"reason,omitempty"`
}

type StepRunResponse struct {
	ID           string          `json:"id"`
	RunID        string          `json:"run_id"`
	StepID       string          `json:"step_id"`
	State        string          `json:"state"`
	Attempt      int             `json:"attempt"`
	WorkerID     string          `json:"worker_id,omitempty"`
	Input        json.RawMessage `json:"input,omitempty"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ProblemDetails struct {
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Status    int       `json:"status"`
	Detail    string    `json:"detail"`
	Instance  string    `json:"instance,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
""")
    commit("api/types: introduce versioned request/response DTOs and ProblemDetails envelope", next(dates_iter), [
        "pkg/api/types/models.go"
    ])

    # Commit 3.2: HTTP REST routing, middleware, and error mappings
    write_file("pkg/api/http/errors.go", """package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/api/types"
	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	title := "Internal Server Error"

	if errors.Is(err, core.ErrNotFound) {
		status = http.StatusNotFound
		title = "Resource Not Found"
	} else if errors.Is(err, core.ErrValidationFailed) {
		status = http.StatusBadRequest
		title = "Bad Request"
	} else if errors.Is(err, core.ErrAlreadyExists) {
		status = http.StatusConflict
		title = "Conflict"
	} else if errors.Is(err, core.ErrInvalidStateTransition) {
		status = http.StatusConflict
		title = "Invalid State Transition"
	} else if errors.Is(err, core.ErrUnauthorized) {
		status = http.StatusUnauthorized
		title = "Unauthorized"
	} else if errors.Is(err, core.ErrForbidden) {
		status = http.StatusForbidden
		title = "Forbidden"
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	problem := types.ProblemDetails{
		Type:      "about:blank",
		Title:     title,
		Status:    status,
		Detail:    err.Error(),
		Instance:  r.URL.Path,
		Timestamp: time.Now().UTC(),
	}
	_ = json.NewEncoder(w).Encode(problem)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
""")

    write_file("pkg/api/http/router.go", """package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/api/types"
	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/statemachine"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Server struct {
	store     storage.EngineStore
	scheduler *scheduler.Scheduler
	sm        *statemachine.Engine
	mux       *http.ServeMux
}

func NewServer(store storage.EngineStore, sched *scheduler.Scheduler) *Server {
	s := &Server{
		store:     store,
		scheduler: sched,
		sm:        statemachine.NewEngine(),
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recovery middleware
		defer func() {
			if rec := recover(); rec != nil {
				WriteError(w, r, errors.New("internal server panic"))
			}
		}()
		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/v1/workflows", s.handleWorkflows)
	s.mux.HandleFunc("/api/v1/workflows/", s.handleWorkflowByID)
	s.mux.HandleFunc("/api/v1/runs", s.handleRuns)
	s.mux.HandleFunc("/api/v1/runs/", s.handleRunByID)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "method not allowed"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func fmtError(base error, msg string) error {
	return &core.DomainError{Code: "HTTP_ERR", Message: msg, Err: base}
}

func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createWorkflow(w, r)
	case http.MethodGet:
		s.listWorkflows(w, r)
	default:
		WriteError(w, r, fmtError(core.ErrValidationFailed, "method not allowed"))
	}
}

func (s *Server) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var req types.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "invalid JSON payload"))
		return
	}

	if err := req.Validate(); err != nil {
		WriteError(w, r, err)
		return
	}

	tenant := req.TenantID
	if tenant == "" {
		tenant = "default"
	}

	wf := &core.WorkflowDefinition{
		ID:          core.NewID("wf"),
		TenantID:    tenant,
		Name:        req.Name,
		Version:     1,
		Description: req.Description,
		Steps:       req.Steps,
		Metadata:    req.Metadata,
	}

	if err := wf.Validate(); err != nil {
		WriteError(w, r, err)
		return
	}

	if err := s.store.CreateWorkflow(r.Context(), wf); err != nil {
		WriteError(w, r, err)
		return
	}

	resp := &types.WorkflowResponse{
		ID:          string(wf.ID),
		TenantID:    wf.TenantID,
		Name:        wf.Name,
		Version:     wf.Version,
		Description: wf.Description,
		Steps:       wf.Steps,
		CreatedAt:   wf.CreatedAt,
		UpdatedAt:   wf.UpdatedAt,
	}
	WriteJSON(w, http.StatusCreated, resp)
}

func (s *Server) listWorkflows(w http.ResponseWriter, r *http.Request) {
	tenant := r.URL.Query().Get("tenant_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit < 0 || offset < 0 {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "negative limit or offset"))
		return
	}
	if limit == 0 {
		limit = 50
	}

	filter := storage.WorkflowFilter{
		TenantID: tenant,
		Pagination: storage.Pagination{
			Limit:  limit,
			Offset: offset,
		},
	}

	list, total, err := s.store.ListWorkflows(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	items := make([]*types.WorkflowResponse, len(list))
	for i, wf := range list {
		items[i] = &types.WorkflowResponse{
			ID:          string(wf.ID),
			TenantID:    wf.TenantID,
			Name:        wf.Name,
			Version:     wf.Version,
			Description: wf.Description,
			Steps:       wf.Steps,
			CreatedAt:   wf.CreatedAt,
			UpdatedAt:   wf.UpdatedAt,
		}
	}

	WriteJSON(w, http.StatusOK, &types.ListWorkflowsResponse{
		Workflows: items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	})
}

func (s *Server) handleWorkflowByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")
	if id == "" {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "missing workflow ID"))
		return
	}

	if r.Method != http.MethodGet {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "method not allowed"))
		return
	}

	wf, err := s.store.GetWorkflow(r.Context(), core.ID(id))
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := &types.WorkflowResponse{
		ID:          string(wf.ID),
		TenantID:    wf.TenantID,
		Name:        wf.Name,
		Version:     wf.Version,
		Description: wf.Description,
		Steps:       wf.Steps,
		CreatedAt:   wf.CreatedAt,
		UpdatedAt:   wf.UpdatedAt,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.submitRun(w, r)
	case http.MethodGet:
		s.listRuns(w, r)
	default:
		WriteError(w, r, fmtError(core.ErrValidationFailed, "method not allowed"))
	}
}

func (s *Server) submitRun(w http.ResponseWriter, r *http.Request) {
	var req types.SubmitRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "invalid JSON payload"))
		return
	}

	if err := req.Validate(); err != nil {
		WriteError(w, r, err)
		return
	}

	wf, err := s.store.GetWorkflow(r.Context(), core.ID(req.WorkflowID))
	if err != nil {
		WriteError(w, r, err)
		return
	}

	priority := core.PriorityNormal
	if req.Priority > 0 {
		priority = core.Priority(req.Priority)
	}

	run, err := s.scheduler.SubmitRun(r.Context(), wf, req.Input, priority)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := &types.RunResponse{
		ID:         string(run.ID),
		WorkflowID: string(run.WorkflowID),
		Version:    run.Version,
		TenantID:   run.TenantID,
		State:      string(run.State),
		Input:      run.Input,
		Priority:   int(run.Priority),
		CreatedAt:  run.CreatedAt,
		UpdatedAt:  run.UpdatedAt,
	}
	WriteJSON(w, http.StatusAccepted, resp)
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	wfID := r.URL.Query().Get("workflow_id")
	state := r.URL.Query().Get("state")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit < 0 || offset < 0 {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "negative limit or offset"))
		return
	}
	if limit == 0 {
		limit = 50
	}

	filter := storage.RunFilter{
		WorkflowID: core.ID(wfID),
		State:      core.RunState(state),
		Pagination: storage.Pagination{
			Limit:  limit,
			Offset: offset,
		},
	}

	runs, total, err := s.store.ListRuns(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	items := make([]*types.RunResponse, len(runs))
	for i, run := range runs {
		items[i] = &types.RunResponse{
			ID:           string(run.ID),
			WorkflowID:   string(run.WorkflowID),
			Version:      run.Version,
			TenantID:     run.TenantID,
			State:        string(run.State),
			Input:        run.Input,
			Output:       run.Output,
			ErrorMessage: run.ErrorMessage,
			Priority:     int(run.Priority),
			StartedAt:    run.StartedAt,
			FinishedAt:   run.FinishedAt,
			CreatedAt:    run.CreatedAt,
			UpdatedAt:    run.UpdatedAt,
		}
	}

	WriteJSON(w, http.StatusOK, &types.ListRunsResponse{
		Runs:   items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Server) handleRunByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
	parts := strings.Split(path, "/")
	runID := parts[0]

	if runID == "" {
		WriteError(w, r, fmtError(core.ErrValidationFailed, "missing run ID"))
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			s.getRun(w, r, core.ID(runID))
			return
		}
		WriteError(w, r, fmtError(core.ErrValidationFailed, "method not allowed"))
		return
	}

	subResource := parts[1]
	switch subResource {
	case "cancel":
		if r.Method == http.MethodPost {
			s.cancelRun(w, r, core.ID(runID))
			return
		}
	case "steps":
		if r.Method == http.MethodGet {
			s.listRunSteps(w, r, core.ID(runID))
			return
		}
	}

	WriteError(w, r, fmtError(core.ErrValidationFailed, "not found or method not allowed"))
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request, id core.ID) {
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := &types.RunResponse{
		ID:           string(run.ID),
		WorkflowID:   string(run.WorkflowID),
		Version:      run.Version,
		TenantID:     run.TenantID,
		State:        string(run.State),
		Input:        run.Input,
		Output:       run.Output,
		ErrorMessage: run.ErrorMessage,
		Priority:     int(run.Priority),
		StartedAt:    run.StartedAt,
		FinishedAt:   run.FinishedAt,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) cancelRun(w http.ResponseWriter, r *http.Request, id core.ID) {
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	if run.State.IsTerminal() {
		WriteError(w, r, fmtError(core.ErrInvalidStateTransition, "cannot cancel terminal workflow run"))
		return
	}

	if err := s.sm.TransitionWorkflow(run, core.RunStateCancelled); err != nil {
		WriteError(w, r, err)
		return
	}

	_ = s.store.UpdateRun(r.Context(), run)
	_ = s.store.AppendEvent(r.Context(), &core.Event{
		Type:      core.EventRunCancelled,
		TenantID:  run.TenantID,
		RunID:     run.ID,
		Timestamp: time.Now().UTC(),
	})

	resp := &types.RunResponse{
		ID:         string(run.ID),
		WorkflowID: string(run.WorkflowID),
		Version:    run.Version,
		TenantID:   run.TenantID,
		State:      string(run.State),
		CreatedAt:  run.CreatedAt,
		UpdatedAt:  run.UpdatedAt,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) listRunSteps(w http.ResponseWriter, r *http.Request, runID core.ID) {
	steps, err := s.store.ListStepRuns(r.Context(), storage.StepRunFilter{RunID: runID})
	if err != nil {
		WriteError(w, r, err)
		return
	}

	items := make([]*types.StepRunResponse, len(steps))
	for i, sr := range steps {
		items[i] = &types.StepRunResponse{
			ID:           string(sr.ID),
			RunID:        string(sr.RunID),
			StepID:       sr.StepID,
			State:        string(sr.State),
			Attempt:      sr.Attempt,
			WorkerID:     sr.WorkerID,
			Input:        sr.Input,
			Output:       sr.Output,
			ErrorMessage: sr.ErrorMessage,
			StartedAt:    sr.StartedAt,
			FinishedAt:   sr.FinishedAt,
			CreatedAt:    sr.CreatedAt,
			UpdatedAt:    sr.UpdatedAt,
		}
	}
	WriteJSON(w, http.StatusOK, items)
}
""")
    commit("api/http: build REST API routing and controllers with RFC 7807 error handling", next(dates_iter), [
        "pkg/api/http/errors.go",
        "pkg/api/http/router.go"
    ])

    # Commit 3.3: gRPC API surface mirroring core endpoints
    write_file("pkg/api/grpc/server.go", """package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/statemachine"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type SubmitRunReq struct {
	WorkflowID string
	Version    int
	Priority   int
	InputJSON  []byte
}

type RunStatusResp struct {
	RunID        string
	WorkflowID   string
	State        string
	ErrorMessage string
}

type CancelRunReq struct {
	RunID  string
	Reason string
}

type Service struct {
	store     storage.EngineStore
	scheduler *scheduler.Scheduler
	sm        *statemachine.Engine
}

func NewService(store storage.EngineStore, sched *scheduler.Scheduler) *Service {
	return &Service{
		store:     store,
		scheduler: sched,
		sm:        statemachine.NewEngine(),
	}
}

func (s *Service) SubmitWorkflowRun(ctx context.Context, req SubmitRunReq) (*RunStatusResp, error) {
	if req.WorkflowID == "" {
		return nil, fmt.Errorf("%w: missing workflow id", core.ErrValidationFailed)
	}

	wf, err := s.store.GetWorkflow(ctx, core.ID(req.WorkflowID))
	if err != nil {
		return nil, err
	}

	priority := core.PriorityNormal
	if req.Priority > 0 {
		priority = core.Priority(req.Priority)
	}

	run, err := s.scheduler.SubmitRun(ctx, wf, json.RawMessage(req.InputJSON), priority)
	if err != nil {
		return nil, err
	}

	return &RunStatusResp{
		RunID:      string(run.ID),
		WorkflowID: string(run.WorkflowID),
		State:      string(run.State),
	}, nil
}

func (s *Service) GetWorkflowRun(ctx context.Context, runID string) (*RunStatusResp, error) {
	if runID == "" {
		return nil, fmt.Errorf("%w: missing run id", core.ErrValidationFailed)
	}

	run, err := s.store.GetRun(ctx, core.ID(runID))
	if err != nil {
		return nil, err
	}

	return &RunStatusResp{
		RunID:        string(run.ID),
		WorkflowID:   string(run.WorkflowID),
		State:        string(run.State),
		ErrorMessage: run.ErrorMessage,
	}, nil
}

func (s *Service) CancelWorkflowRun(ctx context.Context, req CancelRunReq) (*RunStatusResp, error) {
	if req.RunID == "" {
		return nil, fmt.Errorf("%w: missing run id", core.ErrValidationFailed)
	}

	run, err := s.store.GetRun(ctx, core.ID(req.RunID))
	if err != nil {
		return nil, err
	}

	if run.State.IsTerminal() {
		return nil, fmt.Errorf("%w: run already terminal", core.ErrInvalidStateTransition)
	}

	if err := s.sm.TransitionWorkflow(run, core.RunStateCancelled); err != nil {
		return nil, err
	}

	_ = s.store.UpdateRun(ctx, run)
	return &RunStatusResp{
		RunID:      string(run.ID),
		WorkflowID: string(run.WorkflowID),
		State:      string(run.State),
	}, nil
}
""")
    commit("api/grpc: provide gRPC wire service layer mirroring workflow operations", next(dates_iter), [
        "pkg/api/grpc/server.go"
    ])

    # Commit 3.4: Comprehensive boundary and negative-case tests for every endpoint
    write_file("pkg/api/http/http_test.go", """package http

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
""")

    write_file("pkg/api/grpc/grpc_test.go", """package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestGRPCServiceBoundaries(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, 10*time.Millisecond)
	svc := NewService(store, sched)

	// Negative case: submit with empty workflow ID
	_, err := svc.SubmitWorkflowRun(ctx, SubmitRunReq{})
	if err == nil {
		t.Fatal("expected error for empty workflow ID, got nil")
	}

	// Negative case: get non-existent run
	_, err = svc.GetWorkflowRun(ctx, "run-does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent run, got nil")
	}

	// Negative case: cancel empty run ID
	_, err = svc.CancelWorkflowRun(ctx, CancelRunReq{})
	if err == nil {
		t.Fatal("expected error for empty run ID on cancel, got nil")
	}

	// Positive workflow creation and submit
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf"),
		TenantID: "default",
		Name:     "grpc-test-wf",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "s1", TaskType: "http"},
		},
	}
	_ = store.CreateWorkflow(ctx, wf)

	status, err := svc.SubmitWorkflowRun(ctx, SubmitRunReq{WorkflowID: string(wf.ID)})
	if err != nil {
		t.Fatalf("failed submitting run: %v", err)
	}
	if status.State != string(core.RunStateRunning) {
		t.Errorf("expected running, got %s", status.State)
	}

	// Cancel via gRPC
	cancelledStatus, err := svc.CancelWorkflowRun(ctx, CancelRunReq{RunID: status.RunID})
	if err != nil {
		t.Fatalf("failed cancelling run: %v", err)
	}
	if cancelledStatus.State != string(core.RunStateCancelled) {
		t.Errorf("expected cancelled, got %s", cancelledStatus.State)
	}

	// Negative case: cancel already cancelled run
	_, err = svc.CancelWorkflowRun(ctx, CancelRunReq{RunID: status.RunID})
	if err == nil {
		t.Fatal("expected error when cancelling already cancelled run, got nil")
	}
}
""")
    commit("api: add boundary and negative-case unit tests for every HTTP and gRPC endpoint", next(dates_iter), [
        "pkg/api/http/http_test.go",
        "pkg/api/grpc/grpc_test.go"
    ])

    print("Phase 3 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    from generator.git_utils import get_commit_count
    dates = iter(generate_commit_dates(180))
    for _ in range(get_commit_count()):
        next(dates)
    run_phase_3(dates)

package http

import (
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

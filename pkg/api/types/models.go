package types

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

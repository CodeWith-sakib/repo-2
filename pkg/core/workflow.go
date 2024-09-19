package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
	BackoffFactor   float64       `json:"backoff_factor"`
	Jitter          bool          `json:"jitter"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
	}
}

type StepDefinition struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TaskType    string          `json:"task_type"`
	DependsOn   []string        `json:"depends_on"`
	Timeout     time.Duration   `json:"timeout"`
	RetryPolicy *RetryPolicy    `json:"retry_policy,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Condition   string          `json:"condition,omitempty"`
}

type WorkflowDefinition struct {
	ID          ID               `json:"id"`
	TenantID    string           `json:"tenant_id"`
	Name        string           `json:"name"`
	Version     int              `json:"version"`
	Description string           `json:"description"`
	Steps       []StepDefinition `json:"steps"`
	Timeout     time.Duration    `json:"timeout"`
	Metadata    Metadata         `json:"metadata,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (w *WorkflowDefinition) Validate() error {
	if w.ID.IsEmpty() {
		return fmt.Errorf("%w: workflow ID must not be empty", ErrValidationFailed)
	}
	if w.Name == "" {
		return fmt.Errorf("%w: workflow name must not be empty", ErrValidationFailed)
	}
	if w.Version <= 0 {
		return fmt.Errorf("%w: workflow version must be positive", ErrValidationFailed)
	}
	if len(w.Steps) == 0 {
		return fmt.Errorf("%w: workflow must have at least one step", ErrValidationFailed)
	}

	stepMap := make(map[string]bool)
	for _, step := range w.Steps {
		if step.ID == "" {
			return fmt.Errorf("%w: step ID must not be empty", ErrValidationFailed)
		}
		if step.TaskType == "" {
			return fmt.Errorf("%w: step %s task type must not be empty", ErrValidationFailed, step.ID)
		}
		if stepMap[step.ID] {
			return fmt.Errorf("%w: duplicate step ID %s", ErrValidationFailed, step.ID)
		}
		stepMap[step.ID] = true
	}

	for _, step := range w.Steps {
		for _, dep := range step.DependsOn {
			if dep == step.ID {
				return fmt.Errorf("%w: step %s cannot depend on itself", ErrValidationFailed, step.ID)
			}
			if !stepMap[dep] {
				return fmt.Errorf("%w: step %s depends on undefined step %s", ErrValidationFailed, step.ID, dep)
			}
		}
	}

	// Validate DAG for cycles
	dag, err := BuildDAG(w.Steps)
	if err != nil {
		return err
	}
	if err := dag.ValidateAcyclic(); err != nil {
		return err
	}

	return nil
}

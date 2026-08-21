package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrValidationPipelineHalted = errors.New("pipeline validation error detected")
)

// PipelineValidationDiagnostic records warnings or validation errors in a DAG configuration.
type PipelineValidationDiagnostic struct {
	StepID   string `json:"step_id"`
	Severity string `json:"severity"` // "ERROR", "WARNING"
	Message  string `json:"message"`
}

// PipelineDryRunValidator analyzes pipeline definitions before queue submission.
type PipelineDryRunValidator struct {
	mu           sync.RWMutex
	knownTaskMap map[string]bool
}

// NewPipelineDryRunValidator initializes a validator with recognized task types.
func NewPipelineDryRunValidator(registeredTaskTypes []string) *PipelineDryRunValidator {
	m := make(map[string]bool)
	for _, t := range registeredTaskTypes {
		m[t] = true
	}
	return &PipelineDryRunValidator{
		knownTaskMap: m,
	}
}

// ValidatePipeline performs structural checks on DAG steps, dependencies, and task handlers.
func (v *PipelineDryRunValidator) ValidatePipeline(ctx context.Context, def *WorkflowDefinition) ([]PipelineValidationDiagnostic, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var diagnostics []PipelineValidationDiagnostic

	if def == nil {
		return nil, fmt.Errorf("%w: nil workflow definition", ErrValidationPipelineHalted)
	}

	stepIDs := make(map[string]bool)
	for _, s := range def.Steps {
		if stepIDs[s.ID] {
			diagnostics = append(diagnostics, PipelineValidationDiagnostic{
				StepID:   s.ID,
				Severity: "ERROR",
				Message:  fmt.Sprintf("duplicate step identifier %s", s.ID),
			})
		}
		stepIDs[s.ID] = true

		if len(v.knownTaskMap) > 0 && !v.knownTaskMap[s.TaskType] {
			diagnostics = append(diagnostics, PipelineValidationDiagnostic{
				StepID:   s.ID,
				Severity: "WARNING",
				Message:  fmt.Sprintf("unrecognized task type %s; fallback handler will be invoked", s.TaskType),
			})
		}
	}

	// Verify all dependencies resolve
	for _, s := range def.Steps {
		for _, dep := range s.DependsOn {
			if !stepIDs[dep] {
				diagnostics = append(diagnostics, PipelineValidationDiagnostic{
					StepID:   s.ID,
					Severity: "ERROR",
					Message:  fmt.Sprintf("depends on non-existent step %s", dep),
				})
			}
		}
	}

	hasError := false
	for _, d := range diagnostics {
		if d.Severity == "ERROR" {
			hasError = true
			break
		}
	}

	if hasError {
		return diagnostics, ErrValidationPipelineHalted
	}
	return diagnostics, nil
}

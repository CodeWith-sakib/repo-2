package core

import (
	"context"
	"fmt"
	"time"
)

// SimulationStepResult records expected behavior and outcome of a simulated step.
type SimulationStepResult struct {
	StepID            string        `json:"step_id"`
	TaskType          string        `json:"task_type"`
	DependsOn         []string      `json:"depends_on"`
	SimulatedDuration time.Duration `json:"simulated_duration"`
	ParametersPassed  int           `json:"parameters_passed"`
	Status            string        `json:"status"`
}

// WorkflowSimulationReport provides static pre-flight validation and execution summary.
type WorkflowSimulationReport struct {
	WorkflowID       ID                     `json:"workflow_id"`
	TotalSteps       int                    `json:"total_steps"`
	ExecutionWaves   int                    `json:"execution_waves"`
	EstimatedDuration time.Duration         `json:"estimated_duration"`
	StepResults      []SimulationStepResult `json:"step_results"`
	ValidationErrors []string               `json:"validation_errors"`
	Passed           bool                   `json:"passed"`
}

// WorkflowDryRunSimulator validates and simulates DAG step dispatch without executing live tasks.
type WorkflowDryRunSimulator struct {
	defaultStepDuration time.Duration
}

// NewWorkflowDryRunSimulator creates a dry-run simulator.
func NewWorkflowDryRunSimulator(stepDuration time.Duration) *WorkflowDryRunSimulator {
	if stepDuration <= 0 {
		stepDuration = 100 * time.Millisecond
	}
	return &WorkflowDryRunSimulator{
		defaultStepDuration: stepDuration,
	}
}

// Simulate runs a pre-flight execution pass on workflow definition.
func (s *WorkflowDryRunSimulator) Simulate(ctx context.Context, def *WorkflowDefinition) (*WorkflowSimulationReport, error) {
	if def == nil {
		return nil, fmt.Errorf("workflow definition is nil")
	}

	report := &WorkflowSimulationReport{
		WorkflowID:       def.ID,
		TotalSteps:       len(def.Steps),
		Passed:           true,
		StepResults:      make([]SimulationStepResult, 0, len(def.Steps)),
		ValidationErrors: make([]string, 0),
	}

	// Validate graph
	if err := def.Validate(); err != nil {
		report.ValidationErrors = append(report.ValidationErrors, err.Error())
		report.Passed = false
		return report, nil
	}

	resolver := NewDynamicDependencyResolver(def.Steps)
	waves, err := resolver.ResolveExecutionWaves()
	if err != nil {
		report.ValidationErrors = append(report.ValidationErrors, err.Error())
		report.Passed = false
		return report, nil
	}

	report.ExecutionWaves = len(waves)
	report.EstimatedDuration = time.Duration(len(waves)) * s.defaultStepDuration

	for _, step := range def.Steps {
		report.StepResults = append(report.StepResults, SimulationStepResult{
			StepID:            step.ID,
			TaskType:          step.TaskType,
			DependsOn:         step.DependsOn,
			SimulatedDuration: s.defaultStepDuration,
			Status:            "SIMULATED_SUCCESS",
		})
	}

	return report, nil
}

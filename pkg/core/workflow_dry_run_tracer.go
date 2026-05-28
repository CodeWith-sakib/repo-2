package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DryRunStepTrace models simulated step execution without side-effects.
type DryRunStepTrace struct {
	StepID            string            `json:"step_id"`
	EstimatedDuration time.Duration     `json:"estimated_duration"`
	InputBindings     map[string]string `json:"input_bindings"`
	SimulatedOutput   map[string]string `json:"simulated_output"`
	Dependencies      []string          `json:"dependencies"`
	MockSuccess       bool              `json:"mock_success"`
}

// WorkflowDryRunReport captures the overall simulated run of a DAG.
type WorkflowDryRunReport struct {
	WorkflowID         string            `json:"workflow_id"`
	TotalEstimatedTime time.Duration     `json:"total_estimated_time"`
	StepTraces         []DryRunStepTrace `json:"step_traces"`
	CriticalPath       []string          `json:"critical_path"`
	HasCycle           bool              `json:"has_cycle"`
}

// WorkflowDryRunTracer analyzes and simulates workflow execution paths.
type WorkflowDryRunTracer struct {
	mu           sync.RWMutex
	defaultDelay time.Duration
}

// NewWorkflowDryRunTracer initializes a dry-run planner.
func NewWorkflowDryRunTracer(defaultStepDelay time.Duration) *WorkflowDryRunTracer {
	if defaultStepDelay <= 0 {
		defaultStepDelay = 100 * time.Millisecond
	}
	return &WorkflowDryRunTracer{
		defaultDelay: defaultStepDelay,
	}
}

// SimulateExecution performs a virtual execution pass over the given step dependency graph.
func (t *WorkflowDryRunTracer) SimulateExecution(ctx context.Context, workflowID string, steps map[string][]string) (*WorkflowDryRunReport, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	report := &WorkflowDryRunReport{
		WorkflowID:   workflowID,
		StepTraces:   make([]DryRunStepTrace, 0, len(steps)),
		CriticalPath: make([]string, 0),
	}

	// Validate against simple self-cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var checkCycle func(node string) bool
	checkCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range steps[node] {
			if !visited[dep] && checkCycle(dep) {
				return true
			} else if recStack[dep] {
				return true
			}
		}
		recStack[node] = false
		return false
	}

	for node := range steps {
		if !visited[node] {
			if checkCycle(node) {
				report.HasCycle = true
				return report, fmt.Errorf("cycle detected involving node %s", node)
			}
		}
	}

	// Build step traces
	var totalTime time.Duration
	for stepID, deps := range steps {
		trace := DryRunStepTrace{
			StepID:            stepID,
			EstimatedDuration: t.defaultDelay,
			Dependencies:      deps,
			MockSuccess:       true,
			InputBindings:     map[string]string{"mode": "dry-run"},
			SimulatedOutput:   map[string]string{"status": "simulated_ok"},
		}
		totalTime += t.defaultDelay
		report.StepTraces = append(report.StepTraces, trace)
		report.CriticalPath = append(report.CriticalPath, stepID)
	}

	report.TotalEstimatedTime = totalTime
	return report, nil
}

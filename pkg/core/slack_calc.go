package core

import (
	"fmt"
	"time"
)

// TaskScheduleInfo contains timing and critical-path information for a workflow step.
type TaskScheduleInfo struct {
	StepID            string
	EstimatedDuration time.Duration
	EarlyStart        time.Duration
	EarlyFinish       time.Duration
	LateStart         time.Duration
	LateFinish        time.Duration
	TotalSlack        time.Duration
	FreeSlack         time.Duration
	IsCritical        bool
}

// CriticalPathEngine calculates forward and backward passes across DAG tasks to compute slack.
type CriticalPathEngine struct {
	dag *DAG
}

// NewCriticalPathEngine initializes a critical path calculator from a DAG.
func NewCriticalPathEngine(dag *DAG) *CriticalPathEngine {
	return &CriticalPathEngine{dag: dag}
}

// CalculateSlack evaluates Total Slack and Free Slack for every step in the DAG.
func (c *CriticalPathEngine) CalculateSlack(durations map[string]time.Duration) (map[string]*TaskScheduleInfo, error) {
	if c.dag == nil {
		return nil, fmt.Errorf("nil dag")
	}

	order, err := c.dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("topological sort failed: %w", err)
	}

	infoMap := make(map[string]*TaskScheduleInfo)
	for _, node := range c.dag.Nodes() {
		d := durations[node.ID]
		if d == 0 {
			d = time.Second // default fallback duration
		}
		infoMap[node.ID] = &TaskScheduleInfo{
			StepID:            node.ID,
			EstimatedDuration: d,
		}
	}

	// 1. Forward Pass (Early Start & Early Finish)
	var maxProjectFinish time.Duration
	for _, stepID := range order {
		info := infoMap[stepID]
		var maxPredFinish time.Duration
		for _, depID := range c.dag.GetDependencies(stepID) {
			if pred, ok := infoMap[depID]; ok {
				if pred.EarlyFinish > maxPredFinish {
					maxPredFinish = pred.EarlyFinish
				}
			}
		}
		info.EarlyStart = maxPredFinish
		info.EarlyFinish = info.EarlyStart + info.EstimatedDuration
		if info.EarlyFinish > maxProjectFinish {
			maxProjectFinish = info.EarlyFinish
		}
	}

	// 2. Backward Pass (Late Start & Late Finish)
	for _, info := range infoMap {
		info.LateFinish = maxProjectFinish
		info.LateStart = info.LateFinish - info.EstimatedDuration
	}

	// Iterate in reverse topological order
	for i := len(order) - 1; i >= 0; i-- {
		stepID := order[i]
		info := infoMap[stepID]

		dependents := c.dag.GetDependents(stepID)
		if len(dependents) > 0 {
			var minSuccLateStart time.Duration = -1
			var minSuccEarlyStart time.Duration = -1

			for _, succID := range dependents {
				succInfo := infoMap[succID]
				if minSuccLateStart == -1 || succInfo.LateStart < minSuccLateStart {
					minSuccLateStart = succInfo.LateStart
				}
				if minSuccEarlyStart == -1 || succInfo.EarlyStart < minSuccEarlyStart {
					minSuccEarlyStart = succInfo.EarlyStart
				}
			}

			if minSuccLateStart != -1 {
				info.LateFinish = minSuccLateStart
				info.LateStart = info.LateFinish - info.EstimatedDuration
			}
			info.FreeSlack = minSuccEarlyStart - info.EarlyFinish
		} else {
			info.FreeSlack = maxProjectFinish - info.EarlyFinish
		}

		info.TotalSlack = info.LateStart - info.EarlyStart
		if info.TotalSlack == 0 {
			info.IsCritical = true
		}
	}

	return infoMap, nil
}

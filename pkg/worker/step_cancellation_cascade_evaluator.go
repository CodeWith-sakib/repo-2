package worker

import (
	"sync"
)

// CancellationCascadeEvaluator identifies downstream step runs to abort when an upstream task fails or is canceled.
type CancellationCascadeEvaluator struct {
	mu sync.RWMutex
}

// NewCancellationCascadeEvaluator creates a new evaluator.
func NewCancellationCascadeEvaluator() *CancellationCascadeEvaluator {
	return &CancellationCascadeEvaluator{}
}

// ComputeAffectedSteps traverses DAG successors of canceled steps to find all invalid downstream nodes.
func (e *CancellationCascadeEvaluator) ComputeAffectedSteps(canceledStep string, adjList map[string][]string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	visited := make(map[string]bool)
	var cascade []string

	var dfs func(node string)
	dfs = func(node string) {
		for _, child := range adjList[node] {
			if !visited[child] {
				visited[child] = true
				cascade = append(cascade, child)
				dfs(child)
			}
		}
	}

	dfs(canceledStep)
	return cascade
}

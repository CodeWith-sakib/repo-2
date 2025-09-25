package core

import (
	"fmt"
	"sort"
)

// OptimizationReport summarizes optimizations applied to a workflow DAG.
type OptimizationReport struct {
	OriginalSteps   int
	OptimizedSteps  int
	FusedSteps      int
	PrunedDeadSteps int
	RemovedEdges    int
}

// DAGOptimizer analyzes and transforms workflow step graphs for minimal execution overhead.
type DAGOptimizer struct{}

// NewDAGOptimizer creates a DAG optimizer instance.
func NewDAGOptimizer() *DAGOptimizer {
	return &DAGOptimizer{}
}

// Optimize runs transitive reduction and dead-step pruning on the provided steps.
func (o *DAGOptimizer) Optimize(steps []StepDefinition) ([]StepDefinition, OptimizationReport, error) {
	report := OptimizationReport{
		OriginalSteps: len(steps),
	}

	if len(steps) == 0 {
		return nil, report, nil
	}

	// 1. Build step map and validate
	stepMap := make(map[string]StepDefinition, len(steps))
	for _, s := range steps {
		stepMap[s.ID] = s
	}

	for _, s := range steps {
		for _, dep := range s.DependsOn {
			if _, exists := stepMap[dep]; !exists {
				return nil, report, fmt.Errorf("step %q depends on non-existent step %q", s.ID, dep)
			}
		}
	}

	// 2. Transitive reduction (remove redundant dependencies: if A->B and B->C and A->C, remove A->C)
	reducedSteps, removedEdges := o.transitiveReduction(steps)
	report.RemovedEdges = removedEdges

	report.OptimizedSteps = len(reducedSteps)
	return reducedSteps, report, nil
}

// transitiveReduction removes redundant edges from the DAG.
func (o *DAGOptimizer) transitiveReduction(steps []StepDefinition) ([]StepDefinition, int) {
	// Build reachability matrix using Floyd-Warshall style BFS
	adj := make(map[string]map[string]bool)
	for _, s := range steps {
		adj[s.ID] = make(map[string]bool)
		for _, dep := range s.DependsOn {
			adj[dep][s.ID] = true // edge from dependency to step
		}
	}

	// Compute transitive closure (is Y reachable from X in 2 or more hops?)
	reachIn2Plus := make(map[string]map[string]bool)
	for _, s := range steps {
		reachIn2Plus[s.ID] = make(map[string]bool)
	}

	for u := range adj {
		for v := range adj[u] {
			// for each neighbor v of u, all reachable nodes from v are reachable from u in >=2 hops
			visited := make(map[string]bool)
			var queue []string
			for w := range adj[v] {
				queue = append(queue, w)
				visited[w] = true
			}
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				reachIn2Plus[u][curr] = true
				for next := range adj[curr] {
					if !visited[next] {
						visited[next] = true
						queue = append(queue, next)
					}
				}
			}
		}
	}

	// Filter direct edges that are also reachable via >= 2 hops
	removedCount := 0
	optimized := make([]StepDefinition, len(steps))
	for i, s := range steps {
		var prunedDeps []string
		for _, dep := range s.DependsOn {
			if reachIn2Plus[dep][s.ID] {
				// Redundant dependency!
				removedCount++
			} else {
				prunedDeps = append(prunedDeps, dep)
			}
		}
		sort.Strings(prunedDeps)
		newStep := s
		newStep.DependsOn = prunedDeps
		optimized[i] = newStep
	}

	return optimized, removedCount
}

// FindRoots returns IDs of steps that have no dependencies.
func (o *DAGOptimizer) FindRoots(steps []StepDefinition) []string {
	var roots []string
	for _, s := range steps {
		if len(s.DependsOn) == 0 {
			roots = append(roots, s.ID)
		}
	}
	sort.Strings(roots)
	return roots
}

// FindTerminals returns IDs of steps that no other step depends upon.
func (o *DAGOptimizer) FindTerminals(steps []StepDefinition) []string {
	isDep := make(map[string]bool)
	for _, s := range steps {
		for _, d := range s.DependsOn {
			isDep[d] = true
		}
	}

	var terminals []string
	for _, s := range steps {
		if !isDep[s.ID] {
			terminals = append(terminals, s.ID)
		}
	}
	sort.Strings(terminals)
	return terminals
}

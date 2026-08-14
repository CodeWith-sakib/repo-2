package core

import (
	"context"
	"sync"
)

// DependencyGraphPruningPlan outlines pruned nodes and edge contractions.
type DependencyGraphPruningPlan struct {
	RetainedNodes  []string            `json:"retained_nodes"`
	PrunedNodes    []string            `json:"pruned_nodes"`
	CompactedEdges map[string][]string `json:"compacted_edges"`
}

// DependencyGraphPruner trims inactive or conditionally skipped subgraphs from DAG execution plans.
type DependencyGraphPruner struct {
	mu sync.RWMutex
}

// NewDependencyGraphPruner creates a DAG compaction utility.
func NewDependencyGraphPruner() *DependencyGraphPruner {
	return &DependencyGraphPruner{}
}

// PruneDisabledSteps removes inactive nodes and links transitive dependencies directly.
func (p *DependencyGraphPruner) PruneDisabledSteps(ctx context.Context, adjList map[string][]string, enabledNodes map[string]bool) *DependencyGraphPruningPlan {
	p.mu.RLock()
	defer p.mu.RUnlock()

	plan := &DependencyGraphPruningPlan{
		RetainedNodes:  make([]string, 0),
		PrunedNodes:    make([]string, 0),
		CompactedEdges: make(map[string][]string),
	}

	for node := range adjList {
		if enabledNodes[node] {
			plan.RetainedNodes = append(plan.RetainedNodes, node)
		} else {
			plan.PrunedNodes = append(plan.PrunedNodes, node)
		}
	}

	// Compact edges: if A -> B -> C and B is disabled, contract edge to A -> C
	for u := range adjList {
		if !enabledNodes[u] {
			continue
		}

		visited := make(map[string]bool)
		var resolvedTargets []string

		var findActiveTargets func(curr string)
		findActiveTargets = func(curr string) {
			for _, nxt := range adjList[curr] {
				if visited[nxt] {
					continue
				}
				visited[nxt] = true
				if enabledNodes[nxt] {
					resolvedTargets = append(resolvedTargets, nxt)
				} else {
					findActiveTargets(nxt)
				}
			}
		}

		findActiveTargets(u)
		plan.CompactedEdges[u] = resolvedTargets
	}

	return plan
}

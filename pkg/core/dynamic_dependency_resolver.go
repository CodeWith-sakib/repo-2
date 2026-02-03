package core

import (
	"fmt"
	"strings"
)

// DynamicDependencyResolver dynamically calculates step execution order and detects cycles in workflow DAGs.
type DynamicDependencyResolver struct {
	steps        map[string]StepDefinition
	adjacency    map[string][]string
	inDegree     map[string]int
}

// NewDynamicDependencyResolver initializes a dynamic dependency resolver from step definitions.
func NewDynamicDependencyResolver(steps []StepDefinition) *DynamicDependencyResolver {
	r := &DynamicDependencyResolver{
		steps:     make(map[string]StepDefinition),
		adjacency: make(map[string][]string),
		inDegree:  make(map[string]int),
	}

	for _, s := range steps {
		r.steps[s.ID] = s
		r.inDegree[s.ID] = 0
		r.adjacency[s.ID] = nil
	}

	for _, s := range steps {
		for _, dep := range s.DependsOn {
			r.adjacency[dep] = append(r.adjacency[dep], s.ID)
			r.inDegree[s.ID]++
		}
	}

	return r
}

// ResolveExecutionWaves groups independent steps into concurrent parallel waves.
func (r *DynamicDependencyResolver) ResolveExecutionWaves() ([][]string, error) {
	inDeg := make(map[string]int, len(r.inDegree))
	for k, v := range r.inDegree {
		inDeg[k] = v
	}

	var waves [][]string
	processedCount := 0

	for {
		var currentWave []string
		for id, deg := range inDeg {
			if deg == 0 {
				currentWave = append(currentWave, id)
			}
		}

		if len(currentWave) == 0 {
			break
		}

		for _, id := range currentWave {
			delete(inDeg, id)
			processedCount++
			for _, neighbor := range r.adjacency[id] {
				inDeg[neighbor]--
			}
		}

		waves = append(waves, currentWave)
	}

	if processedCount < len(r.steps) {
		var remaining []string
		for id := range inDeg {
			remaining = append(remaining, id)
		}
		return nil, fmt.Errorf("cyclic dependency detected among steps: %s", strings.Join(remaining, ", "))
	}

	return waves, nil
}

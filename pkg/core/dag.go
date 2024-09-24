package core

import (
	"fmt"
	"sort"
)

type DAG struct {
	nodes        map[string]StepDefinition
	adjacency    map[string][]string // stepID -> list of dependents (steps that depend on stepID)
	dependencies map[string][]string // stepID -> list of prerequisites (steps that stepID depends on)
}

func BuildDAG(steps []StepDefinition) (*DAG, error) {
	d := &DAG{
		nodes:        make(map[string]StepDefinition),
		adjacency:    make(map[string][]string),
		dependencies: make(map[string][]string),
	}

	for _, s := range steps {
		d.nodes[s.ID] = s
		d.adjacency[s.ID] = make([]string, 0)
		d.dependencies[s.ID] = make([]string, len(s.DependsOn))
		copy(d.dependencies[s.ID], s.DependsOn)
	}

	for _, s := range steps {
		for _, dep := range s.DependsOn {
			d.adjacency[dep] = append(d.adjacency[dep], s.ID)
		}
	}

	return d, nil
}

func (d *DAG) GetNode(id string) (StepDefinition, bool) {
	node, exists := d.nodes[id]
	return node, exists
}

func (d *DAG) Nodes() []StepDefinition {
	list := make([]StepDefinition, 0, len(d.nodes))
	for _, n := range d.nodes {
		list = append(list, n)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

func (d *DAG) GetDependencies(id string) []string {
	deps := d.dependencies[id]
	out := make([]string, len(deps))
	copy(out, deps)
	return out
}

func (d *DAG) GetDependents(id string) []string {
	deps := d.adjacency[id]
	out := make([]string, len(deps))
	copy(out, deps)
	return out
}

func (d *DAG) ValidateAcyclic() error {
	inDegree := make(map[string]int)
	for id := range d.nodes {
		inDegree[id] = len(d.dependencies[id])
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visitedCount := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visitedCount++

		for _, dep := range d.adjacency[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if visitedCount != len(d.nodes) {
		return fmt.Errorf("%w: visited %d nodes out of %d total", ErrCycleDetected, visitedCount, len(d.nodes))
	}
	return nil
}

func (d *DAG) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for id := range d.nodes {
		inDegree[id] = len(d.dependencies[id])
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	result := make([]string, 0, len(d.nodes))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		result = append(result, curr)

		nextNodes := make([]string, 0)
		for _, dep := range d.adjacency[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				nextNodes = append(nextNodes, dep)
			}
		}
		sort.Strings(nextNodes)
		queue = append(queue, nextNodes...)
	}

	if len(result) != len(d.nodes) {
		return nil, ErrCycleDetected
	}
	return result, nil
}

func (d *DAG) RootNodes() []string {
	roots := make([]string, 0)
	for id, deps := range d.dependencies {
		if len(deps) == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	return roots
}

func (d *DAG) LeafNodes() []string {
	leaves := make([]string, 0)
	for id, adjs := range d.adjacency {
		if len(adjs) == 0 {
			leaves = append(leaves, id)
		}
	}
	sort.Strings(leaves)
	return leaves
}

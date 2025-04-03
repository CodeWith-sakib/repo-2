package core

import (
	"fmt"
	"sort"
)

type SubgraphPartitioner struct {
	dag *DAG
}

func NewSubgraphPartitioner(dag *DAG) *SubgraphPartitioner {
	return &SubgraphPartitioner{dag: dag}
}

// PartitionByCriticality groups steps into independent parallel stages or execution tiers.
func (p *SubgraphPartitioner) ExecutionTiers() ([][]string, error) {
	inDegree := make(map[string]int)
	nodes := p.dag.Nodes()
	for _, n := range nodes {
		inDegree[n.ID] = len(p.dag.GetDependencies(n.ID))
	}

	var tiers [][]string
	for len(inDegree) > 0 {
		var currentTier []string
		for id, deg := range inDegree {
			if deg == 0 {
				currentTier = append(currentTier, id)
			}
		}

		if len(currentTier) == 0 {
			return nil, fmt.Errorf("cycle detected or unresolved dependency")
		}

		sort.Strings(currentTier)
		tiers = append(tiers, currentTier)

		for _, id := range currentTier {
			delete(inDegree, id)
			for _, dep := range p.dag.GetDependents(id) {
				if _, ok := inDegree[dep]; ok {
					inDegree[dep]--
				}
			}
		}
	}

	return tiers, nil
}

// ReachableSubgraphs returns all downstream dependent nodes reachable from a given start node.
func (p *SubgraphPartitioner) ReachableDownstream(startNode string) ([]string, error) {
	if _, ok := p.dag.GetNode(startNode); !ok {
		return nil, fmt.Errorf("node %q not found in DAG", startNode)
	}

	visited := make(map[string]bool)
	var queue []string
	queue = append(queue, startNode)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, child := range p.dag.GetDependents(curr) {
			if !visited[child] {
				visited[child] = true
				queue = append(queue, child)
			}
		}
	}

	result := make([]string, 0, len(visited))
	for id := range visited {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

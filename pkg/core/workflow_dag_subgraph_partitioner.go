package core

import (
	"sync"
)

// BreadthClusterPartition models a group of workflow nodes partitioned for distributed chunk execution.
type BreadthClusterPartition struct {
	ClusterIndex  int      `json:"cluster_index"`
	NodeIDs       []string `json:"node_ids"`
	InternalEdges int      `json:"internal_edges"`
}

// BreadthClusterPartitioner splits an adjacency list into chunked execution clusters using topological traversal.
type BreadthClusterPartitioner struct {
	mu           sync.RWMutex
	maxPerSubDAG int
}

// NewBreadthClusterPartitioner initializes a breadth-first cluster partitioner.
func NewBreadthClusterPartitioner(maxPerSubDAG int) *BreadthClusterPartitioner {
	if maxPerSubDAG <= 0 {
		maxPerSubDAG = 10
	}
	return &BreadthClusterPartitioner{
		maxPerSubDAG: maxPerSubDAG,
	}
}

// PartitionByBreadth splits nodes into clusters.
func (p *BreadthClusterPartitioner) PartitionByBreadth(adjList map[string][]string) ([]BreadthClusterPartition, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(adjList) == 0 {
		return nil, nil
	}

	inDegree := make(map[string]int)
	for u := range adjList {
		if _, ok := inDegree[u]; !ok {
			inDegree[u] = 0
		}
		for _, v := range adjList[u] {
			inDegree[v]++
		}
	}

	queue := make([]string, 0)
	for u, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, u)
		}
	}

	var partitions []BreadthClusterPartition
	currentNodes := make([]string, 0, p.maxPerSubDAG)
	pIdx := 0

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currentNodes = append(currentNodes, curr)

		if len(currentNodes) >= p.maxPerSubDAG {
			partitions = append(partitions, BreadthClusterPartition{
				ClusterIndex: pIdx,
				NodeIDs:      currentNodes,
			})
			pIdx++
			currentNodes = make([]string, 0, p.maxPerSubDAG)
		}

		for _, v := range adjList[curr] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if len(currentNodes) > 0 {
		partitions = append(partitions, BreadthClusterPartition{
			ClusterIndex: pIdx,
			NodeIDs:      currentNodes,
		})
	}

	return partitions, nil
}

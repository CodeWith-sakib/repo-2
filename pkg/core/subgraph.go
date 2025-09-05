package core

import (
	"fmt"
	"sort"
)

// SubgraphPartitioner decomposes a large workflow DAG into independent parallelisable sub-graphs
// using a topological layering (critical-path level assignment) algorithm. Steps at the same
// level have no mutual dependencies and can execute concurrently.

// LayerAssignment maps each step ID to its parallelism level (0 = root layer).
type LayerAssignment map[string]int

// SubgraphPartition is a named group of steps assigned the same level.
type SubgraphPartition struct {
	Level int
	Steps []string
}

// PartitionDAG assigns each step to a parallelism level and returns ordered partitions.
func PartitionDAG(steps []StepDefinition) ([]SubgraphPartition, error) {
	// Build dependency lookup
	deps := make(map[string][]string, len(steps))
	stepSet := make(map[string]bool, len(steps))
	for _, s := range steps {
		deps[s.ID] = s.DependsOn
		stepSet[s.ID] = true
	}

	// Validate all deps exist
	for _, s := range steps {
		for _, d := range s.DependsOn {
			if !stepSet[d] {
				return nil, fmt.Errorf("step %q depends on unknown step %q", s.ID, d)
			}
		}
	}

	// Compute level via longest-path BFS
	level := make(map[string]int, len(steps))
	// Explicitly initialize all step levels to 0
	for _, s := range steps {
		level[s.ID] = 0
	}
	inDegree := make(map[string]int, len(steps))
	adjacency := make(map[string][]string, len(steps))

	for _, s := range steps {
		inDegree[s.ID] = len(s.DependsOn)
		for _, d := range s.DependsOn {
			adjacency[d] = append(adjacency[d], s.ID)
		}
	}

	queue := []string{}
	for _, s := range steps {
		if inDegree[s.ID] == 0 {
			queue = append(queue, s.ID)
		}
	}
	sort.Strings(queue)

	processed := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		processed++

		for _, succ := range adjacency[curr] {
			if level[curr]+1 > level[succ] {
				level[succ] = level[curr] + 1
			}
			inDegree[succ]--
			if inDegree[succ] == 0 {
				queue = append(queue, succ)
			}
		}
	}

	if processed != len(steps) {
		return nil, fmt.Errorf("%w: cycle detected — only processed %d/%d steps", ErrCycleDetected, processed, len(steps))
	}

	// Group by level
	byLevel := make(map[int][]string)
	maxLevel := 0
	for id, l := range level {
		byLevel[l] = append(byLevel[l], id)
		if l > maxLevel {
			maxLevel = l
		}
	}

	partitions := make([]SubgraphPartition, 0, maxLevel+1)
	for l := 0; l <= maxLevel; l++ {
		group := byLevel[l]
		sort.Strings(group)
		partitions = append(partitions, SubgraphPartition{Level: l, Steps: group})
	}

	return partitions, nil
}

// CriticalPathLength computes the length (number of levels) of the critical path through the DAG.
func CriticalPathLength(partitions []SubgraphPartition) int {
	return len(partitions)
}

// Bottlenecks returns steps that are the sole member of their level (serial bottlenecks).
func Bottlenecks(partitions []SubgraphPartition) []string {
	var result []string
	for _, p := range partitions {
		if len(p.Steps) == 1 {
			result = append(result, p.Steps[0])
		}
	}
	return result
}

// ParallelismFactor returns the average number of steps per level (higher = more parallelism).
func ParallelismFactor(partitions []SubgraphPartition) float64 {
	if len(partitions) == 0 {
		return 0
	}
	total := 0
	for _, p := range partitions {
		total += len(p.Steps)
	}
	return float64(total) / float64(len(partitions))
}

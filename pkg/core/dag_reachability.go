package core

import (
	"fmt"
	"sort"
)

// BitSet is a compact dynamic bit array representation.
type BitSet []uint64

// NewBitSet creates a bit set capable of holding n bits.
func NewBitSet(n int) BitSet {
	words := (n + 63) / 64
	return make(BitSet, words)
}

// Set sets the bit at index i.
func (b BitSet) Set(i int) {
	word := i / 64
	bit := uint(i % 64)
	if word < len(b) {
		b[word] |= (1 << bit)
	}
}

// Test checks whether the bit at index i is set.
func (b BitSet) Test(i int) bool {
	word := i / 64
	bit := uint(i % 64)
	if word < len(b) {
		return (b[word] & (1 << bit)) != 0
	}
	return false
}

// Or performs bitwise OR with other in place.
func (b BitSet) Or(other BitSet) {
	minLen := len(b)
	if len(other) < minLen {
		minLen = len(other)
	}
	for i := 0; i < minLen; i++ {
		b[i] |= other[i]
	}
}

// Count returns the number of set bits (Hamming weight).
func (b BitSet) Count() int {
	total := 0
	for _, word := range b {
		v := word
		// Brian Kernighan bit count
		for v != 0 {
			v &= v - 1
			total++
		}
	}
	return total
}

// DAGReachabilityIndex provides O(1) topological reachability queries between any pair of steps.
type DAGReachabilityIndex struct {
	idToIndex map[string]int
	indexToID []string
	ancestors []BitSet // ancestors[i] has bit j set if j is an ancestor of i
}

// BuildReachabilityIndex constructs a bitset reachability matrix from a DAG definition.
func BuildReachabilityIndex(steps []StepDefinition) (*DAGReachabilityIndex, error) {
	n := len(steps)
	idToIndex := make(map[string]int, n)
	indexToID := make([]string, n)

	// Sort step IDs for deterministic indexing
	sortedSteps := make([]StepDefinition, n)
	copy(sortedSteps, steps)
	sort.Slice(sortedSteps, func(i, j int) bool {
		return sortedSteps[i].ID < sortedSteps[j].ID
	})

	for i, s := range sortedSteps {
		idToIndex[s.ID] = i
		indexToID[i] = s.ID
	}

	ancestors := make([]BitSet, n)
	for i := range ancestors {
		ancestors[i] = NewBitSet(n)
	}

	// Compute in topological sort order
	dag, err := BuildDAG(steps)
	if err != nil {
		return nil, fmt.Errorf("failed to build DAG: %w", err)
	}

	topoOrder, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("topological sort failed: %w", err)
	}

	for _, id := range topoOrder {
		idx := idToIndex[id]
		for _, dep := range dag.GetDependencies(id) {
			depIdx := idToIndex[dep]
			// Direct dependency is an ancestor
			ancestors[idx].Set(depIdx)
			// Transitive ancestors of dependency are also ancestors
			ancestors[idx].Or(ancestors[depIdx])
		}
	}

	return &DAGReachabilityIndex{
		idToIndex: idToIndex,
		indexToID: indexToID,
		ancestors: ancestors,
	}, nil
}

// CanReach returns true if source can reach target (i.e. source is an ancestor of target).
func (idx *DAGReachabilityIndex) CanReach(sourceID, targetID string) bool {
	srcIdx, srcOk := idx.idToIndex[sourceID]
	tgtIdx, tgtOk := idx.idToIndex[targetID]
	if !srcOk || !tgtOk {
		return false
	}
	return idx.ancestors[tgtIdx].Test(srcIdx)
}

// AncestorCount returns the total number of ancestor steps for a step.
func (idx *DAGReachabilityIndex) AncestorCount(stepID string) int {
	i, ok := idx.idToIndex[stepID]
	if !ok {
		return 0
	}
	return idx.ancestors[i].Count()
}

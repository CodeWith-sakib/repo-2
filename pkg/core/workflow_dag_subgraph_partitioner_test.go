package core

import (
	"testing"
)

func TestBreadthClusterPartitioner(t *testing.T) {
	partitioner := NewBreadthClusterPartitioner(2)

	adj := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
		"D": {},
	}

	partitions, err := partitioner.PartitionByBreadth(adj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(partitions) < 2 {
		t.Errorf("expected at least 2 partitions, got %d", len(partitions))
	}

	totalNodes := 0
	for _, p := range partitions {
		totalNodes += len(p.NodeIDs)
	}
	if totalNodes != 4 {
		t.Errorf("expected 4 nodes partitioned, got %d", totalNodes)
	}
}

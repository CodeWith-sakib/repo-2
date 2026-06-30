package events

import (
	"testing"
)

func TestPartitionHashRing(t *testing.T) {
	ring := NewPartitionHashRing(20)

	ring.AddNode("node-alpha")
	ring.AddNode("node-beta")
	ring.AddNode("node-gamma")

	node1, err := ring.GetNode("tenant-user-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node1 == "" {
		t.Error("expected non-empty node")
	}

	// Consistency test
	node1Again, _ := ring.GetNode("tenant-user-1234")
	if node1 != node1Again {
		t.Errorf("consistent hashing failed: got %s vs %s", node1, node1Again)
	}

	// Remove node
	ring.RemoveNode(node1)
	nodeAfter, err := ring.GetNode("tenant-user-1234")
	if err != nil || nodeAfter == node1 {
		t.Errorf("expected reassignment away from removed node, got %s", nodeAfter)
	}
}

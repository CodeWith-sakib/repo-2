package postgres

import (
	"testing"
)

func TestDeadlockDetectorGraph(t *testing.T) {
	graph := NewDeadlockDetectorGraph()

	// TX 1 waits on TX 2, TX 2 waits on TX 3
	graph.AddWaitEdge(1, 2, "row-a")
	graph.AddWaitEdge(2, 3, "row-b")

	cycles := graph.DetectCycles()
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %d", len(cycles))
	}

	// TX 3 waits on TX 1 -> cycle!
	graph.AddWaitEdge(3, 1, "row-c")
	cycles = graph.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("expected cycle detected, got none")
	}

	// Resolve deadlock by committing TX 1
	graph.ClearTXEdges(1)
	cyclesAfter := graph.DetectCycles()
	if len(cyclesAfter) != 0 {
		t.Errorf("expected cycle cleared after TX 1 resolution, got %d", len(cyclesAfter))
	}
}

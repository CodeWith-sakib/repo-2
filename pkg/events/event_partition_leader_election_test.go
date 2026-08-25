package events

import (
	"context"
	"testing"
	"time"
)

func TestPartitionLeaderElection(t *testing.T) {
	election := NewPartitionLeaderElection()

	if !election.TryAcquireLease(context.Background(), 0, "node-1", 100*time.Millisecond) {
		t.Fatal("node-1 should acquire lease")
	}

	if election.TryAcquireLease(context.Background(), 0, "node-2", 100*time.Millisecond) {
		t.Fatal("node-2 should not acquire active lease")
	}

	leader, ok := election.GetLeader(0)
	if !ok || leader != "node-1" {
		t.Fatalf("expected node-1 as leader, got %s (ok=%v)", leader, ok)
	}

	time.Sleep(120 * time.Millisecond)
	// Now node-2 can claim expired lease
	if !election.TryAcquireLease(context.Background(), 0, "node-2", 100*time.Millisecond) {
		t.Fatal("node-2 should acquire lease after expiration")
	}
}

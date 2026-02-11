package postgres

import (
	"context"
	"testing"
	"time"
)

func TestReplicaLagDetector(t *testing.T) {
	detector := NewReplicaLagDetector(nil)

	res, err := detector.CheckReplicationLag(context.Background(), 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error checking lag: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 replicas, got %d", len(res))
	}

	for _, r := range res {
		if !r.IsHealthy {
			t.Errorf("expected replica %s to report healthy", r.ClientAddr)
		}
		if r.ReplayLag <= 0 {
			t.Errorf("expected positive replay lag for %s", r.ClientAddr)
		}
	}
}

package postgres

import (
	"context"
	"testing"
)

func TestConnectionAffinityRouter(t *testing.T) {
	primary := DBConnectionNode{ID: "pg-primary", Host: "10.0.0.1", Port: 5432, IsPrimary: true}
	replicas := []DBConnectionNode{
		{ID: "pg-replica-1", Host: "10.0.0.2", Port: 5432, IsPrimary: false},
		{ID: "pg-replica-2", Host: "10.0.0.3", Port: 5432, IsPrimary: false},
	}

	router := NewConnectionAffinityRouter(primary, replicas)

	// Writes must always route to primary
	targetWrite := router.RouteTarget(context.Background(), "tenant-a", true)
	if targetWrite.ID != "pg-primary" {
		t.Errorf("write must go to primary, got %s", targetWrite.ID)
	}

	// Read routes consistently to replica
	r1 := router.RouteTarget(context.Background(), "tenant-a", false)
	r2 := router.RouteTarget(context.Background(), "tenant-a", false)
	if r1.ID != r2.ID || r1.IsPrimary {
		t.Errorf("expected sticky replica routing, got %s vs %s", r1.ID, r2.ID)
	}
}

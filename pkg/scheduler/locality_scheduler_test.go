package scheduler

import (
	"testing"
)

func TestLocalityScheduler_SelectsHostLocal(t *testing.T) {
	sched := NewLocalityScheduler()

	// 3 nodes in different locations
	sched.RegisterNode(NodeTopology{
		NodeID: "node-1", Host: "host-A", Rack: "rack-1", Zone: "us-east-1a", Capacity: 10,
	})
	sched.RegisterNode(NodeTopology{
		NodeID: "node-2", Host: "host-B", Rack: "rack-1", Zone: "us-east-1a", Capacity: 10,
	})
	sched.RegisterNode(NodeTopology{
		NodeID: "node-3", Host: "host-C", Rack: "rack-2", Zone: "us-east-1b", Capacity: 10,
	})

	// Prefer host-A -> should choose node-1
	pref := LocalityPreference{
		PreferredHost: "host-A",
	}

	best, err := sched.SelectNode(pref)
	if err != nil {
		t.Fatalf("select node failed: %v", err)
	}
	if best.NodeID != "node-1" {
		t.Errorf("expected node-1 (host-local), got %s", best.NodeID)
	}

	// Prefer rack-1 -> node-1 and node-2 are candidates; if node-1 is heavily loaded, pick node-2
	sched.RegisterNode(NodeTopology{
		NodeID: "node-1", Host: "host-A", Rack: "rack-1", Zone: "us-east-1a", ActiveTasks: 9, Capacity: 10,
	})
	rackPref := LocalityPreference{
		PreferredRack: "rack-1",
	}
	bestRack, err := sched.SelectNode(rackPref)
	if err != nil {
		t.Fatalf("select rack failed: %v", err)
	}
	if bestRack.NodeID != "node-2" {
		t.Errorf("expected node-2 (unloaded rack-local), got %s", bestRack.NodeID)
	}
}

func TestLocalityScheduler_AllFull(t *testing.T) {
	sched := NewLocalityScheduler()
	sched.RegisterNode(NodeTopology{
		NodeID: "n-full", Capacity: 5, ActiveTasks: 5,
	})

	_, err := sched.SelectNode(LocalityPreference{})
	if err == nil {
		t.Error("expected error when all nodes saturated")
	}
}

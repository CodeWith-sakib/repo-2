package postgres

import (
	"testing"
	"time"
)

func TestReplicaRouter_Routing(t *testing.T) {
	primary := DatabaseNode{ID: "primary", DSN: "postgres://primary:5432/db"}
	router := NewReplicaRouter(primary, DefaultRouterConfig())

	rep1 := DatabaseNode{ID: "replica-1", DSN: "postgres://rep1:5432/db"}
	rep2 := DatabaseNode{ID: "replica-2", DSN: "postgres://rep2:5432/db"}
	router.AddReplica(rep1)
	router.AddReplica(rep2)

	// Write query -> should route to primary
	node, err := router.RouteQuery("INSERT INTO users (id) VALUES (1)", "session-1")
	if err != nil || node.ID != "primary" {
		t.Fatalf("expected write to route to primary, got %s (err=%v)", node.ID, err)
	}

	// Immediate read query from same session -> should route to primary (sticky!)
	readSticky, err := router.RouteQuery("SELECT * FROM users WHERE id = 1", "session-1")
	if err != nil || readSticky.ID != "primary" {
		t.Errorf("expected sticky read to route to primary, got %s", readSticky.ID)
	}

	// Read query from fresh session -> should route to replica
	readFresh, err := router.RouteQuery("SELECT * FROM users WHERE id = 1", "session-2")
	if err != nil || (readFresh.ID != "replica-1" && readFresh.ID != "replica-2") {
		t.Errorf("expected fresh read to route to replica, got %s", readFresh.ID)
	}

	// Locking read (SELECT FOR UPDATE) -> must route to primary
	lockRead, err := router.RouteQuery("SELECT * FROM orders WHERE id = 5 FOR UPDATE", "session-3")
	if err != nil || lockRead.ID != "primary" {
		t.Errorf("expected SELECT FOR UPDATE to route to primary, got %s", lockRead.ID)
	}
}

func TestReplicaRouter_FallbackOnReplicaLag(t *testing.T) {
	primary := DatabaseNode{ID: "primary"}
	cfg := RouterConfig{
		MaxLagBytes:    1024,
		StickyDuration: time.Second,
	}
	router := NewReplicaRouter(primary, cfg)

	rep := DatabaseNode{ID: "rep-laggy", LagBytes: 10000} // exceeds 1024 max lag
	router.AddReplica(rep)

	// Read query should fallback to primary because replica is lagging
	node, _ := router.RouteQuery("SELECT 1", "s1")
	if node.ID != "primary" {
		t.Errorf("expected fallback to primary on lagging replica, got %s", node.ID)
	}
}

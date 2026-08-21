package events

import (
	"context"
	"testing"
)

func TestFanoutBroadcastRouter(t *testing.T) {
	router := NewFanoutBroadcastRouter()

	router.RegisterDestination("audit.orders", FanoutDestination{ID: "es-indexer", QueueLen: 500})
	router.RegisterDestination("audit.orders", FanoutDestination{ID: "s3-archiver", QueueLen: 1000})

	if router.DestinationCount("audit.orders") != 2 {
		t.Fatalf("expected 2 destinations, got %d", router.DestinationCount("audit.orders"))
	}

	dests := router.RouteFanout(context.Background(), "audit.orders")
	if len(dests) != 2 {
		t.Errorf("expected 2 routed destinations, got %d", len(dests))
	}
}

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/events"
)

func TestDefinitionCacheHitMissInvalidate(t *testing.T) {
	c := NewDefinitionCache(10, 50*time.Millisecond)

	wf := &core.WorkflowDefinition{
		ID:      core.NewID("wf"),
		Name:    "cached-wf",
		Version: 1,
	}

	// Initial Miss
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected cache miss, got hit")
	}

	// Put and Hit
	c.Put(wf)
	cached, ok := c.Get(wf.ID, wf.Version)
	if !ok || cached.Name != "cached-wf" {
		t.Fatalf("expected cache hit, got ok=%v", ok)
	}

	// Explicit invalidation by version
	c.InvalidateVersion(wf.ID, wf.Version)
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected miss after InvalidateVersion, got hit")
	}

	// Put and Invalidate by ID
	c.Put(wf)
	c.Invalidate(wf.ID)
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected miss after Invalidate, got hit")
	}
}

func TestDefinitionCacheEventBusInvalidation(t *testing.T) {
	c := NewDefinitionCache(10, time.Minute)
	bus := events.NewBus()
	c.AttachEventBus(bus)

	wf := &core.WorkflowDefinition{
		ID:      core.NewID("wf-event"),
		Name:    "event-wf",
		Version: 1,
	}

	c.Put(wf)
	if _, ok := c.Get(wf.ID, wf.Version); !ok {
		t.Fatal("expected cache hit before event")
	}

	// Fire EventWorkflowUpdated
	bus.Publish(context.Background(), &core.Event{
		Type:  core.EventWorkflowUpdated,
		RunID: wf.ID, // identifier in event
	})

	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected cache invalidated after EventWorkflowUpdated, but item was still cached")
	}
}

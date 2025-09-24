package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	topic := "workflow.completed"
	sub, err := bus.Subscribe(topic, 5)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	evt := &core.Event{
		ID:        core.NewID("evt"),
		Type:      core.EventType(topic),
		TenantID:  "tenant-a",
		Timestamp: time.Now(),
	}

	if err := bus.Publish(evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case received := <-sub.Ch:
		if received.ID != evt.ID {
			t.Errorf("expected event ID %s, got %s", evt.ID, received.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}

	pub, del, drop, dlq := bus.Metrics()
	if pub != 1 || del != 1 || drop != 0 || dlq != 0 {
		t.Errorf("metrics mismatch: pub=%d del=%d drop=%d dlq=%d", pub, del, drop, dlq)
	}

	bus.Unsubscribe(sub)
}

func TestEventBus_DeadLetter(t *testing.T) {
	bus := NewEventBus(5)
	defer bus.Close()

	evt := &core.Event{
		ID:        core.NewID("evt-dlq"),
		Type:      "unsubscribed.topic",
		TenantID:  "tenant-b",
		Timestamp: time.Now(),
	}

	if err := bus.Publish(evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case dlqEvt := <-bus.DeadLetterChannel():
		if dlqEvt.ID != evt.ID {
			t.Errorf("expected DLQ event ID %s, got %s", evt.ID, dlqEvt.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for DLQ event")
	}

	_, _, _, dlq := bus.Metrics()
	if dlq != 1 {
		t.Errorf("expected 1 dead letter, got %d", dlq)
	}
}

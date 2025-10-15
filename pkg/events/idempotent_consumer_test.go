package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestIdempotentEventConsumer_DuplicateSuppression(t *testing.T) {
	handledCount := 0
	handler := func(ctx context.Context, event *core.Event) error {
		handledCount++
		return nil
	}

	consumer := NewIdempotentEventConsumer(handler, time.Hour)

	evt := &core.Event{
		ID:        core.NewID("evt-001"),
		Type:      "order.created",
		Timestamp: time.Now(),
	}

	ctx := context.Background()

	// 1st delivery -> processed
	if err := consumer.HandleEvent(ctx, evt); err != nil {
		t.Fatalf("first delivery failed: %v", err)
	}

	// 2nd delivery of same event -> duplicate, suppressed!
	if err := consumer.HandleEvent(ctx, evt); err != nil {
		t.Fatalf("second delivery failed: %v", err)
	}

	// 3rd delivery of same event -> duplicate, suppressed!
	if err := consumer.HandleEvent(ctx, evt); err != nil {
		t.Fatalf("third delivery failed: %v", err)
	}

	if handledCount != 1 {
		t.Errorf("expected handler called exactly once, got %d", handledCount)
	}

	total, processed, dups, _ := consumer.Metrics()
	if total != 3 || processed != 1 || dups != 2 {
		t.Errorf("metrics mismatch: total=%d processed=%d dups=%d", total, processed, dups)
	}
}

func TestIdempotentEventConsumer_RetryOnError(t *testing.T) {
	attempts := 0
	handler := func(ctx context.Context, event *core.Event) error {
		attempts++
		if attempts == 1 {
			return fmt.Errorf("transient network failure")
		}
		return nil
	}

	consumer := NewIdempotentEventConsumer(handler, time.Hour)

	evt := &core.Event{
		ID:        core.NewID("evt-002"),
		Type:      "payment.process",
		Timestamp: time.Now(),
	}

	ctx := context.Background()

	// 1st attempt fails
	if err := consumer.HandleEvent(ctx, evt); err == nil {
		t.Fatal("expected error on 1st attempt")
	}

	// 2nd attempt should be allowed because 1st errored
	if err := consumer.HandleEvent(ctx, evt); err != nil {
		t.Fatalf("expected 2nd attempt to succeed: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts executed, got %d", attempts)
	}
}

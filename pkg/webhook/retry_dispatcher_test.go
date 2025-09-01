package webhook

import (
	"context"
	"testing"
	"time"
)

func TestRetryDispatcher_SuccessOnSecondAttempt(t *testing.T) {
	callCount := 0
	deliverFn := func(ctx context.Context, url string, payload []byte) (int, error) {
		callCount++
		if callCount < 2 {
			return 503, nil // first attempt fails
		}
		return 200, nil // second succeeds
	}

	dispatcher := NewRetryDispatcher(deliverFn)
	cfg := WebhookEndpointConfig{
		URL:            "https://hooks.example.com/event",
		MaxAttempts:    3,
		InitialBackoff: 5 * time.Millisecond,
		BackoffFactor:  2.0,
		MaxBackoff:     50 * time.Millisecond,
		Timeout:        100 * time.Millisecond,
	}

	result := dispatcher.Dispatch(context.Background(), cfg, []byte(`{"event":"test"}`))

	if !result.Delivered {
		t.Errorf("expected successful delivery, got %+v", result)
	}
	if len(result.Attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(result.Attempts))
	}
}

func TestRetryDispatcher_ExhaustsAllAttempts(t *testing.T) {
	deliverFn := func(ctx context.Context, url string, payload []byte) (int, error) {
		return 500, nil // always fails
	}

	dispatcher := NewRetryDispatcher(deliverFn)
	cfg := WebhookEndpointConfig{
		URL:            "https://failing.example.com",
		MaxAttempts:    3,
		InitialBackoff: 2 * time.Millisecond,
		BackoffFactor:  1.5,
		MaxBackoff:     10 * time.Millisecond,
		Timeout:        50 * time.Millisecond,
	}

	result := dispatcher.Dispatch(context.Background(), cfg, []byte(`{}`))

	if result.Delivered {
		t.Error("expected delivery failure")
	}
	if len(result.Attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", len(result.Attempts))
	}
	if result.FinalError == "" {
		t.Error("expected non-empty final error message")
	}
}

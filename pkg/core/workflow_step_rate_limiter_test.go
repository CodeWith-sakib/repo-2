package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStepRateLimiter(t *testing.T) {
	limiter := NewStepRateLimiter()
	limiter.SetLimit("http_call", 5)

	for i := 0; i < 5; i++ {
		if !limiter.Allow("http_call") {
			t.Fatalf("first 5 calls should succeed, failed at %d", i)
		}
	}

	if limiter.Allow("http_call") {
		t.Fatal("6th call should be throttled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx, "http_call")
	if !errors.Is(err, ErrStepRateLimitExceeded) {
		t.Fatalf("expected ErrStepRateLimitExceeded, got %v", err)
	}
}

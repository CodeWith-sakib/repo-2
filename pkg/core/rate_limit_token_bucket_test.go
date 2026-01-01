package core

import (
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	limiter := NewTokenBucketLimiter(10.0, 5.0) // 10 tokens/sec, capacity 5

	now := time.Now().UTC()
	// Consume all initial tokens
	for i := 0; i < 5; i++ {
		if !limiter.AllowN(now, 1.0) {
			t.Fatalf("expected token %d to be allowed", i)
		}
	}

	// 6th should be rejected immediately at same timestamp
	if limiter.AllowN(now, 1.0) {
		t.Error("expected 6th token to be rejected")
	}

	// Advance time by 0.5s -> 5 tokens replenished
	future := now.Add(500 * time.Millisecond)
	if !limiter.AllowN(future, 4.0) {
		t.Error("expected 4 tokens allowed after 0.5s refill")
	}

	if limiter.AllowN(future, 2.0) {
		t.Error("expected insufficient tokens for 2 tokens")
	}
}

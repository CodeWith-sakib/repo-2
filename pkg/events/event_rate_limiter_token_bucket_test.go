package events

import (
	"testing"
	"time"
)

func TestEventTokenBucketLimiter(t *testing.T) {
	limiter := NewEventTokenBucketLimiter(10, 10)

	if !limiter.Allow(5) {
		t.Error("consuming 5 of 10 tokens should succeed")
	}
	if !limiter.Allow(5) {
		t.Error("consuming remaining 5 tokens should succeed")
	}
	if limiter.Allow(1) {
		t.Error("consuming from exhausted bucket should fail")
	}

	time.Sleep(200 * time.Millisecond) // refills ~2 tokens
	if !limiter.Allow(1) {
		t.Error("consuming after refill should succeed")
	}
}

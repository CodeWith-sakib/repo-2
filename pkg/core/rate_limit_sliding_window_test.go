package core

import (
	"testing"
	"time"
)

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := NewSlidingWindowLimiter(3, 100*time.Millisecond)

	now := time.Now().UTC()
	if !limiter.AllowAt(now) {
		t.Fatal("expected first request allowed")
	}
	if !limiter.AllowAt(now.Add(10 * time.Millisecond)) {
		t.Fatal("expected second request allowed")
	}
	if !limiter.AllowAt(now.Add(20 * time.Millisecond)) {
		t.Fatal("expected third request allowed")
	}

	// 4th within 100ms should be rejected
	if limiter.AllowAt(now.Add(30 * time.Millisecond)) {
		t.Error("expected fourth request within window to be rejected")
	}

	// Advance past 100ms -> oldest should have slid out
	if !limiter.AllowAt(now.Add(120 * time.Millisecond)) {
		t.Error("expected request allowed after window slides past first timestamp")
	}
}

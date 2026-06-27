package core

import (
	"testing"
	"time"
)

func TestRetryJitterAllocator(t *testing.T) {
	alloc := NewRetryJitterAllocator(100*time.Millisecond, 2*time.Second, FullJitter)

	for i := 0; i < 5; i++ {
		d := alloc.ComputeDelay(i)
		if d < 0 || d > 2*time.Second {
			t.Errorf("delay out of bounds: %v", d)
		}
	}

	allocEqual := NewRetryJitterAllocator(100*time.Millisecond, 2*time.Second, EqualJitter)
	dEqual := allocEqual.ComputeDelay(2)
	if dEqual < 200*time.Millisecond {
		t.Errorf("equal jitter should have half baseline, got %v", dEqual)
	}
}

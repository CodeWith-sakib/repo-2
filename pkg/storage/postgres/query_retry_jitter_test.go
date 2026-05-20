package postgres

import (
	"testing"
	"time"
)

func TestFullJitterBackoff(t *testing.T) {
	b := NewFullJitterBackoff(10*time.Millisecond, 200*time.Millisecond)

	for attempt := 0; attempt < 5; attempt++ {
		dur := b.ComputeBackoff(attempt)
		if dur < 0 {
			t.Errorf("backoff duration cannot be negative: %v", dur)
		}
		if dur > 200*time.Millisecond {
			t.Errorf("backoff duration exceeded cap: %v", dur)
		}
	}
}

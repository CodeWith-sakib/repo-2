package core

import (
	"testing"
	"time"
)

func TestLeakyBucketSmoothing(t *testing.T) {
	b := NewLeakyBucket(10, 100) // leaks 100 per sec

	if !b.Add(5) {
		t.Fatal("expected adding 5 to succeed")
	}
	if !b.Add(5) {
		t.Fatal("expected adding another 5 to succeed")
	}
	if b.Add(1) {
		t.Fatal("expected capacity overflow")
	}

	time.Sleep(50 * time.Millisecond) // leaks ~5
	if !b.Add(3) {
		t.Fatal("expected adding 3 to succeed after leak")
	}
}

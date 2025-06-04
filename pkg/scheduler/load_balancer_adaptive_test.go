package scheduler

import (
	"testing"
)

func TestAdaptiveLoadBalancer(t *testing.T) {
	lb := NewAdaptiveLoadBalancer()
	if lb.NextIndex(3) != 0 {
		t.Error("expected 0")
	}
	if lb.NextIndex(3) != 1 {
		t.Error("expected 1")
	}
}

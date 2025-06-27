package scheduler

import "testing"

func TestFairShareDecay(t *testing.T) {
	fsd := NewFairShareDecay(0.5)
	fsd.Record("t1", 10.0)
	fsd.Decay()
	if fsd.Get("t1") != 5.0 {
		t.Errorf("expected 5.0, got %v", fsd.Get("t1"))
	}
}

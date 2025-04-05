package scheduler

import (
	"testing"
)

func TestAffinityMatcher(t *testing.T) {
	matcher := NewAffinityMatcher()

	w1 := &WorkerCapability{
		WorkerID:  "w1",
		Labels:    map[string]string{"zone": "us-east-1a", "gpu": "true"},
		BusySlots: 1,
		MaxSlots:  4,
	}
	w2 := &WorkerCapability{
		WorkerID:  "w2",
		Labels:    map[string]string{"zone": "us-west-1a"},
		BusySlots: 0,
		MaxSlots:  4,
	}

	aff := &NodeAffinity{
		RequiredLabels: map[string]string{"gpu": "true"},
		PreferredKeys:  []string{"zone"},
	}

	if !matcher.Matches(w1, aff) {
		t.Error("expected w1 to match gpu affinity")
	}
	if matcher.Matches(w2, aff) {
		t.Error("expected w2 not to match gpu affinity")
	}

	best := matcher.SelectBestWorker([]*WorkerCapability{w1, w2}, aff)
	if best == nil || best.WorkerID != "w1" {
		t.Errorf("expected w1 selected, got %v", best)
	}
}

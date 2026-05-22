package worker

import (
	"testing"
)

func TestTenantAffinityLoadBalancer(t *testing.T) {
	workers := []string{"node-01", "node-02", "node-03"}
	lb := NewTenantAffinityLoadBalancer(workers)

	// Consistent mapping
	w1 := lb.SelectWorker("tenant-alpha")
	w2 := lb.SelectWorker("tenant-alpha")
	if w1 != w2 {
		t.Errorf("expected consistent worker assignment: %s vs %s", w1, w2)
	}

	// Distinct keys route across cluster
	wBeta := lb.SelectWorker("tenant-beta")
	if wBeta == "" {
		t.Error("expected non-empty worker assignment")
	}
}

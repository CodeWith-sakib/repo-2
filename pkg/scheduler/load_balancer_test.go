package scheduler

import (
	"testing"
)

func TestWorkerLoadBalancer(t *testing.T) {
	lb := NewWorkerLoadBalancer()
	lb.RegisterWorker("w1", 2)
	lb.RegisterWorker("w2", 4)

	sel1, ok := lb.SelectWorker()
	if !ok {
		t.Fatal("expected worker selected")
	}

	sel2, ok := lb.SelectWorker()
	if !ok {
		t.Fatal("expected second worker selected")
	}

	if sel1 == "" || sel2 == "" {
		t.Error("worker ID should not be empty")
	}

	lb.ReleaseTask(sel1)
}

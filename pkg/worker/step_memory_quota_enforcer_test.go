package worker

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryQuotaEnforcer(t *testing.T) {
	enforcer := NewMemoryQuotaEnforcer(1000)

	err := enforcer.AllocateQuota(context.Background(), "step-1", 400)
	if err != nil {
		t.Fatalf("unexpected error allocating step-1: %v", err)
	}

	err = enforcer.AllocateQuota(context.Background(), "step-2", 400)
	if err != nil {
		t.Fatalf("unexpected error allocating step-2: %v", err)
	}

	err = enforcer.AllocateQuota(context.Background(), "step-3", 300)
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}

	enforcer.ReleaseQuota("step-1")
	err = enforcer.AllocateQuota(context.Background(), "step-3", 300)
	if err != nil {
		t.Fatalf("unexpected error allocating step-3 after release: %v", err)
	}

	alloc, _ := enforcer.CurrentAllocations()
	if alloc != 700 {
		t.Errorf("expected 700 bytes allocated, got %d", alloc)
	}
}

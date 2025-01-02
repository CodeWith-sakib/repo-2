package scheduler

import (
	"context"
	"testing"
)

func TestConcurrencyGovernorAcquireRelease(t *testing.T) {
	gov := NewConcurrencyGovernor(TenantLimit{
		MaxConcurrent: 2,
		RatePerSecond: 100,
		BurstCapacity: 10,
	})

	ctx := context.Background()
	tenant := "org-alpha"

	if err := gov.Acquire(ctx, tenant); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if err := gov.Acquire(ctx, tenant); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Third acquire should violate max concurrency 2
	if err := gov.Acquire(ctx, tenant); err == nil {
		t.Fatal("expected max concurrency error, got nil")
	}

	gov.Release(tenant)
	if act := gov.GetActive(tenant); act != 1 {
		t.Fatalf("expected 1 active after release, got %d", act)
	}

	if err := gov.Acquire(ctx, tenant); err != nil {
		t.Fatalf("acquire after release failed: %v", err)
	}
}

package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkflowSemaphorePool(t *testing.T) {
	pool := NewWorkflowSemaphorePool()
	pool.SetTenantLimit("tenant-1", 2)

	if !pool.TryAcquire("tenant-1") {
		t.Fatal("first acquire should succeed")
	}
	if !pool.TryAcquire("tenant-1") {
		t.Fatal("second acquire should succeed")
	}
	if pool.TryAcquire("tenant-1") {
		t.Fatal("third acquire should fail (limit 2)")
	}

	// Test acquire timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := pool.Acquire(ctx, "tenant-1")
	if !errors.Is(err, ErrSemaphoreAcquisitionTimeout) {
		t.Fatalf("expected ErrSemaphoreAcquisitionTimeout, got %v", err)
	}

	pool.Release("tenant-1")
	if pool.ActiveCount("tenant-1") != 1 {
		t.Errorf("expected 1 active after release, got %d", pool.ActiveCount("tenant-1"))
	}
}

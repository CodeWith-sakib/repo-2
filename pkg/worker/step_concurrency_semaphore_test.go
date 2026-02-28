package worker

import (
	"context"
	"testing"
	"time"
)

func TestKeyedConcurrencySemaphore(t *testing.T) {
	sem := NewKeyedConcurrencySemaphore(2)

	if !sem.TryAcquire("tenant-a") {
		t.Fatal("expected slot 1 allowed for tenant-a")
	}
	if !sem.TryAcquire("tenant-a") {
		t.Fatal("expected slot 2 allowed for tenant-a")
	}
	// 3rd rejected
	if sem.TryAcquire("tenant-a") {
		t.Error("expected slot 3 rejected for tenant-a")
	}

	// tenant-b should still have separate quota
	if !sem.TryAcquire("tenant-b") {
		t.Error("expected tenant-b to acquire independent slot")
	}

	// Release tenant-a slot and re-acquire
	sem.Release("tenant-a")
	if !sem.TryAcquire("tenant-a") {
		t.Error("expected slot available after release")
	}

	// Test timed acquire with timeout
	ctx := context.Background()
	err := sem.Acquire(ctx, "tenant-a", 20*time.Millisecond)
	if err == nil {
		t.Error("expected timeout acquiring saturated slot, got nil")
	}
}

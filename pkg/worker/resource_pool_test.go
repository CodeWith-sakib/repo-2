package worker

import (
	"testing"
)

func TestMultiDimensionalResourcePool_AllocationAndRelease(t *testing.T) {
	capacity := ResourceVector{
		CPUMillicores: 4000, // 4 cores
		MemoryMB:      8192, // 8 GB
		GPUCount:      2,    // 2 GPUs
	}

	pool, err := NewMultiDimensionalResourcePool(capacity)
	if err != nil {
		t.Fatalf("create pool failed: %v", err)
	}

	// Allocate task 1
	req1 := ResourceVector{CPUMillicores: 2000, MemoryMB: 4096, GPUCount: 1}
	lease1, err := pool.TryAllocate("t1", req1)
	if err != nil {
		t.Fatalf("allocate t1 failed: %v", err)
	}

	cpu, mem, gpu := pool.Utilization()
	if cpu != 50.0 || mem != 50.0 || gpu != 50.0 {
		t.Errorf("expected 50%% utilization across all, got cpu=%.1f mem=%.1f gpu=%.1f", cpu, mem, gpu)
	}

	// Allocate task 2 (fits remaining 2000m, 4096MB, 1 GPU)
	req2 := ResourceVector{CPUMillicores: 2000, MemoryMB: 4096, GPUCount: 1}
	lease2, err := pool.TryAllocate("t2", req2)
	if err != nil {
		t.Fatalf("allocate t2 failed: %v", err)
	}

	// Allocate task 3 (exceeds CPU/GPU capacity -> should fail)
	req3 := ResourceVector{CPUMillicores: 1000, MemoryMB: 1024, GPUCount: 0}
	_, err = pool.TryAllocate("t3", req3)
	if err == nil {
		t.Error("expected error when capacity exhausted")
	}

	// Release task 1
	if !pool.Release(lease1.LeaseID) {
		t.Fatal("release lease1 failed")
	}

	// Now task 3 should fit
	lease3, err := pool.TryAllocate("t3", req3)
	if err != nil {
		t.Fatalf("allocate t3 failed after release: %v", err)
	}

	pool.Release(lease2.LeaseID)
	pool.Release(lease3.LeaseID)

	cpu, mem, gpu = pool.Utilization()
	if cpu != 0.0 || mem != 0.0 || gpu != 0.0 {
		t.Errorf("expected 0%% utilization after all releases, got cpu=%.1f mem=%.1f gpu=%.1f", cpu, mem, gpu)
	}
}

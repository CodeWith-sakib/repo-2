package scheduler

import (
	"testing"
)

func TestTenantQuotaManager_AcquireAndRelease(t *testing.T) {
	qm := NewTenantQuotaManager()
	qm.SetLimit("tenant-corp", ResourceVector{
		CPUMillicores: 4000,
		MemoryMB:      8192,
		GPUUnits:      2,
	})

	req1 := ResourceVector{CPUMillicores: 2000, MemoryMB: 4096, GPUUnits: 1}
	if err := qm.TryAcquire("tenant-corp", req1); err != nil {
		t.Fatalf("first acquire should succeed: %v", err)
	}

	// Exceeds remaining GPU
	req2 := ResourceVector{CPUMillicores: 1000, MemoryMB: 1024, GPUUnits: 2}
	if err := qm.TryAcquire("tenant-corp", req2); err == nil {
		t.Fatalf("expected error exceeding GPU limit, but got none")
	}

	// Release req1
	qm.Release("tenant-corp", req1)
	usage := qm.GetUsage("tenant-corp")
	if usage.CPUMillicores != 0 || usage.MemoryMB != 0 || usage.GPUUnits != 0 {
		t.Errorf("expected 0 usage after release, got %+v", usage)
	}
}

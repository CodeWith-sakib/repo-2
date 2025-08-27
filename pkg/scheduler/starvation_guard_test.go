package scheduler

import (
	"testing"
	"time"
)

func TestStarvationGuard_BoostTriggered(t *testing.T) {
	boosted := map[string]int{}
	boostFn := func(tenantID string, boost int) {
		boosted[tenantID] += boost
	}

	guard := NewStarvationGuard(50*time.Millisecond, 10, boostFn)

	guard.RecordEnqueue("tenant-a")
	guard.RecordEnqueue("tenant-b")
	guard.RecordExecution("tenant-b") // b is serviced immediately

	time.Sleep(60 * time.Millisecond)

	result := guard.CheckAndBoost()

	if len(result) != 1 || result[0] != "tenant-a" {
		t.Errorf("expected only tenant-a to be boosted, got %v", result)
	}
	if boosted["tenant-a"] != 10 {
		t.Errorf("expected boost magnitude 10 for tenant-a, got %d", boosted["tenant-a"])
	}
	if boosted["tenant-b"] != 0 {
		t.Errorf("tenant-b should not be boosted, got %d", boosted["tenant-b"])
	}
}

func TestStarvationGuard_StatusReport(t *testing.T) {
	guard := NewStarvationGuard(time.Minute, 5, func(_ string, _ int) {})
	guard.RecordEnqueue("tenant-x")
	s := guard.Status()
	if s == "" {
		t.Error("status report should not be empty")
	}
}

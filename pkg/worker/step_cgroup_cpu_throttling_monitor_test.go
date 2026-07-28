package worker

import (
	"testing"
	"time"
)

func TestCgroupCPUThrottlingMonitor(t *testing.T) {
	mon := NewCgroupCPUThrottlingMonitor(0.20)

	m1 := mon.RecordCFSStats(100, 5, 50000000)
	if m1.IsCritical {
		t.Error("5% throttling should not be critical")
	}

	m2 := mon.RecordCFSStats(100, 25, 250000000)
	if !m2.IsCritical {
		t.Error("25% throttling should be critical")
	}
	if m2.TimeThrottled != 250*time.Millisecond {
		t.Errorf("expected 250ms throttled, got %v", m2.TimeThrottled)
	}
}

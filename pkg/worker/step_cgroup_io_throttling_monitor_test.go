package worker

import (
	"testing"
)

func TestCgroupIOThrottlingMonitor(t *testing.T) {
	monitor := NewCgroupIOThrottlingMonitor(10*1024*1024, 1000) // 10MB/s, 1000 IOPS

	m1 := monitor.RecordIOObservation(2*1024*1024, 1*1024*1024, 200, 100)
	if m1.IsThrottled {
		t.Error("3MB/s and 300 IOPS should not trigger throttle warning")
	}

	m2 := monitor.RecordIOObservation(6*1024*1024, 3*1024*1024, 500, 400) // 9MB/s >= 8.5MB/s
	if !m2.IsThrottled {
		t.Error("9MB/s should trigger throttle warning")
	}
}

package worker

import (
	"sync"
	"time"
)

// PressureLevel indicates PSI memory stall severity.
type PressureLevel string

const (
	PressureNone     PressureLevel = "NONE"
	PressureModerate PressureLevel = "MODERATE"
	PressureCritical PressureLevel = "CRITICAL"
)

// PSIMemoryMetrics records Linux pressure stall information (PSI) metrics.
type PSIMemoryMetrics struct {
	SomeAvg10   float64       `json:"some_avg10"`
	SomeAvg60   float64       `json:"some_avg60"`
	FullAvg10   float64       `json:"full_avg10"`
	Level       PressureLevel `json:"level"`
	ObservedAt  time.Time     `json:"observed_at"`
}

// CgroupMemoryPressureWatcher monitors Linux PSI memory stall times.
type CgroupMemoryPressureWatcher struct {
	mu           sync.RWMutex
	warnStallPct float64
	critStallPct float64
}

// NewCgroupMemoryPressureWatcher initializes a PSI memory pressure monitor.
func NewCgroupMemoryPressureWatcher(warnPct, critPct float64) *CgroupMemoryPressureWatcher {
	if warnPct <= 0.0 {
		warnPct = 10.0 // 10% stall time
	}
	if critPct <= warnPct {
		critPct = 30.0 // 30% stall time
	}
	return &CgroupMemoryPressureWatcher{
		warnStallPct: warnPct,
		critStallPct: critPct,
	}
}

// RecordPSIObservation categorizes memory stall intensity into pressure levels.
func (w *CgroupMemoryPressureWatcher) RecordPSIObservation(someAvg10, someAvg60, fullAvg10 float64) PSIMemoryMetrics {
	w.mu.Lock()
	defer w.mu.Unlock()

	level := PressureNone
	if fullAvg10 >= w.critStallPct || someAvg10 >= w.critStallPct {
		level = PressureCritical
	} else if someAvg10 >= w.warnStallPct {
		level = PressureModerate
	}

	return PSIMemoryMetrics{
		SomeAvg10:  someAvg10,
		SomeAvg60:  someAvg60,
		FullAvg10:  fullAvg10,
		Level:      level,
		ObservedAt: time.Now(),
	}
}

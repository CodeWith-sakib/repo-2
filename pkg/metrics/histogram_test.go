package metrics

import (
	"testing"
	"time"
)

func TestLatencyHistogramPercentiles(t *testing.T) {
	hist := NewLatencyHistogram(100)

	for i := 1; i <= 100; i++ {
		hist.Record(time.Duration(i) * time.Millisecond)
	}

	p50 := hist.Percentile(50)
	if p50 != 50*time.Millisecond {
		t.Errorf("expected p50 of 50ms, got %v", p50)
	}

	p99 := hist.Percentile(99)
	if p99 != 99*time.Millisecond {
		t.Errorf("expected p99 of 99ms, got %v", p99)
	}

	mean := hist.Mean()
	if mean != 50*time.Millisecond && mean != 50*time.Millisecond+500*time.Microsecond {
		t.Errorf("unexpected mean: %v", mean)
	}
}

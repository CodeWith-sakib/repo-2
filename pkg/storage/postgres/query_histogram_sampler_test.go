package postgres

import (
	"testing"
	"time"
)

func TestQueryLatencyHistogram(t *testing.T) {
	hist := NewQueryLatencyHistogram(100)

	for i := 1; i <= 100; i++ {
		hist.RecordSample(time.Duration(i) * time.Millisecond)
	}

	if hist.SampleCount() != 100 {
		t.Fatalf("expected 100 samples, got %d", hist.SampleCount())
	}

	p50 := hist.Percentile(0.50)
	if p50 < 45*time.Millisecond || p50 > 55*time.Millisecond {
		t.Errorf("expected ~50ms for P50, got %v", p50)
	}

	p99 := hist.Percentile(0.99)
	if p99 < 95*time.Millisecond {
		t.Errorf("expected ~99ms for P99, got %v", p99)
	}
}

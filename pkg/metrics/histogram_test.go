package metrics

import (
	"testing"
)

func TestExplicitHistogram_ObserveAndQuantile(t *testing.T) {
	bounds := []float64{10, 25, 50, 100, 250, 500, 1000}
	h := NewExplicitHistogram("latency_ms", bounds, map[string]string{"service": "api"})

	// Observe 100 values: 1..100
	for i := 1; i <= 100; i++ {
		h.Observe(float64(i))
	}

	buckets, sum, count := h.Snapshot()
	if count != 100 {
		t.Errorf("expected 100 observations, got %d", count)
	}
	if sum != 5050 {
		t.Errorf("expected sum 5050, got %.1f", sum)
	}
	if len(buckets) == 0 {
		t.Error("expected non-empty buckets")
	}

	// p50 of 1..100 should be around 50
	p50 := h.Quantile(0.5)
	if p50 < 45 || p50 > 55 {
		t.Errorf("p50 out of expected range [45,55]: %.2f", p50)
	}

	// p95 should be around 95
	p95 := h.Quantile(0.95)
	if p95 < 90 || p95 > 100 {
		t.Errorf("p95 out of expected range [90,100]: %.2f", p95)
	}

	mean := h.Mean()
	if mean < 50.0 || mean > 51.0 {
		t.Errorf("mean should be ~50.5, got %.2f", mean)
	}

	if h.String() == "" {
		t.Error("String() should not be empty")
	}
}

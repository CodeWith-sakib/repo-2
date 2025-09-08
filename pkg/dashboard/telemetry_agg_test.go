package dashboard

import (
	"testing"
	"time"
)

func TestTelemetryAggregator_WindowedAggregation(t *testing.T) {
	agg := NewTelemetryAggregator(10000)
	now := time.Now()

	// Ingest 10 points over 10 seconds
	for i := 0; i < 10; i++ {
		agg.Ingest(TelemetryDataPoint{
			MetricName: "api.latency_ms",
			TenantID:   "corp-a",
			Value:      float64(10 + i*5),
			At:         now.Add(time.Duration(i) * time.Second),
		})
	}

	w := AggregationWindow{
		Start:    now,
		End:      now.Add(10 * time.Second),
		Interval: 5 * time.Second,
	}

	series, err := agg.Aggregate("api.latency_ms", "corp-a", w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if series == nil {
		t.Fatal("expected non-nil series")
	}
	if len(series.Buckets) == 0 {
		t.Error("expected non-empty buckets")
	}

	// Total count across all buckets should equal 10
	total := 0
	for _, b := range series.Buckets {
		total += b.Count
	}
	if total != 10 {
		t.Errorf("expected 10 total points, got %d", total)
	}
}

func TestTelemetryAggregator_CapacityTrim(t *testing.T) {
	agg := NewTelemetryAggregator(5)
	now := time.Now()

	for i := 0; i < 10; i++ {
		agg.Ingest(TelemetryDataPoint{
			MetricName: "x",
			TenantID:   "t",
			Value:      float64(i),
			At:         now,
		})
	}

	if agg.PointCount() > 5 {
		t.Errorf("expected buffer capped at 5, got %d", agg.PointCount())
	}
}

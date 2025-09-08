package dashboard

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// TelemetryDataPoint holds a single metric observation with metadata.
type TelemetryDataPoint struct {
	MetricName string
	TenantID   string
	Value      float64
	At         time.Time
	Tags       map[string]string
}

// AggregationWindow defines the time range for metric aggregation.
type AggregationWindow struct {
	Start    time.Time
	End      time.Time
	Interval time.Duration
}

// AggregatedSeries is the output of a time-windowed aggregation.
type AggregatedSeries struct {
	MetricName string
	TenantID   string
	Buckets    []MetricBucket
}

// MetricBucket holds aggregated stats for a time interval.
type MetricBucket struct {
	Start time.Time
	End   time.Time
	Count int
	Sum   float64
	Min   float64
	Max   float64
	Mean  float64
	P95   float64
}

// TelemetryAggregator buffers telemetry data points and computes windowed aggregations.
type TelemetryAggregator struct {
	mu        sync.Mutex
	points    []TelemetryDataPoint
	maxPoints int
}

// NewTelemetryAggregator creates an aggregator with a bounded buffer.
func NewTelemetryAggregator(maxPoints int) *TelemetryAggregator {
	if maxPoints <= 0 {
		maxPoints = 100000
	}
	return &TelemetryAggregator{maxPoints: maxPoints}
}

// Ingest adds a data point to the buffer.
func (a *TelemetryAggregator) Ingest(p TelemetryDataPoint) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.points = append(a.points, p)

	// Trim if over capacity (remove oldest)
	if len(a.points) > a.maxPoints {
		a.points = a.points[len(a.points)-a.maxPoints:]
	}
}

// Aggregate computes time-bucketed aggregations for a metric and tenant over a window.
func (a *TelemetryAggregator) Aggregate(metricName, tenantID string, w AggregationWindow) (*AggregatedSeries, error) {
	if w.Interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}
	if w.Start.After(w.End) {
		return nil, fmt.Errorf("window start must be before end")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Filter matching points
	matched := []TelemetryDataPoint{}
	for _, p := range a.points {
		if p.MetricName == metricName &&
			(tenantID == "" || p.TenantID == tenantID) &&
			!p.At.Before(w.Start) && !p.At.After(w.End) {
			matched = append(matched, p)
		}
	}

	// Sort by time
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].At.Before(matched[j].At)
	})

	// Create time buckets
	numBuckets := int(w.End.Sub(w.Start)/w.Interval) + 1
	if numBuckets > 10000 {
		numBuckets = 10000
	}

	buckets := make([]MetricBucket, numBuckets)
	for i := range buckets {
		bs := w.Start.Add(time.Duration(i) * w.Interval)
		be := bs.Add(w.Interval)
		buckets[i] = MetricBucket{Start: bs, End: be, Min: 1e18, Max: -1e18}
	}

	// Assign points to buckets
	for _, p := range matched {
		idx := int(p.At.Sub(w.Start) / w.Interval)
		if idx < 0 {
			idx = 0
		}
		if idx >= numBuckets {
			idx = numBuckets - 1
		}
		b := &buckets[idx]
		b.Count++
		b.Sum += p.Value
		if p.Value < b.Min {
			b.Min = p.Value
		}
		if p.Value > b.Max {
			b.Max = p.Value
		}
	}

	// Compute mean for each bucket, reset empty-bucket sentinels
	for i := range buckets {
		if buckets[i].Count == 0 {
			buckets[i].Min = 0
			buckets[i].Max = 0
		} else {
			buckets[i].Mean = buckets[i].Sum / float64(buckets[i].Count)
		}
	}

	return &AggregatedSeries{
		MetricName: metricName,
		TenantID:   tenantID,
		Buckets:    buckets,
	}, nil
}

// PointCount returns total buffered points.
func (a *TelemetryAggregator) PointCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.points)
}

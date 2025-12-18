package postgres

import (
	"context"
	"testing"
)

func TestStatsCollectorMock(t *testing.T) {
	collector := NewStatsCollector(nil)

	metrics, err := collector.CollectSlowQueries(context.Background(), 1.0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(metrics) != 2 {
		t.Fatalf("expected 2 mock metrics, got %d", len(metrics))
	}

	if metrics[0].QueryID != "q_workflow_runs_by_tenant" {
		t.Errorf("unexpected query ID: %s", metrics[0].QueryID)
	}

	if metrics[0].MeanTimeMs <= 0 {
		t.Errorf("expected positive mean time, got %f", metrics[0].MeanTimeMs)
	}
}

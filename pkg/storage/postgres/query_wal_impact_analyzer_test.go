package postgres

import (
	"testing"
)

func TestQueryWALImpactAnalyzer(t *testing.T) {
	analyzer := NewQueryWALImpactAnalyzer(1024) // 1KB segments for quick trigger

	for i := 0; i < 10; i++ {
		analyzer.RecordWrite(200)
	}

	m := analyzer.GetMetrics()
	if m.TotalOperations != 10 {
		t.Errorf("expected 10 operations, got %d", m.TotalOperations)
	}
	if m.TotalWALBytes != 10*(200+64) {
		t.Errorf("expected %d total wal bytes, got %d", 10*(200+64), m.TotalWALBytes)
	}
	if m.EstimatedFsyncs < 2 {
		t.Errorf("expected >= 2 fsyncs, got %d", m.EstimatedFsyncs)
	}
	if m.SummaryString() == "" {
		t.Errorf("expected non-empty summary string")
	}
}

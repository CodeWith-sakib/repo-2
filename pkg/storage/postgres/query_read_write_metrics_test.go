package postgres

import (
	"testing"
)

func TestReadWriteTrafficMetrics(t *testing.T) {
	metrics := NewReadWriteTrafficMetrics()

	metrics.RecordMasterQuery()
	metrics.RecordReplicaQuery()
	metrics.RecordReplicaQuery()
	metrics.RecordReplicaQuery()

	if metrics.MasterQueries.Load() != 1 {
		t.Errorf("expected 1 master query, got %d", metrics.MasterQueries.Load())
	}
	if metrics.ReplicaQueries.Load() != 3 {
		t.Errorf("expected 3 replica queries, got %d", metrics.ReplicaQueries.Load())
	}

	ratio := metrics.ReadRatioPct()
	if ratio != 75.0 {
		t.Errorf("expected 75%% read ratio, got %f", ratio)
	}
}

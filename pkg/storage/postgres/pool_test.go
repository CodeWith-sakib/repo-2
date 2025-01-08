package postgres

import (
	"testing"
)

func TestPoolMonitorStats(t *testing.T) {
	// Dummy test with nil db Stats mock
	pm := &PoolMonitor{
		isHealthy: true,
	}
	pm.RecordQuery()
	pm.RecordQuery()

	if pm.queryCount != 2 {
		t.Errorf("expected 2 queries recorded, got %d", pm.queryCount)
	}
	if !pm.IsHealthy() {
		t.Error("expected pool monitor healthy")
	}
}

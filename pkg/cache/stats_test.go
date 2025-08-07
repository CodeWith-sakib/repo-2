package cache

import (
	"testing"
)

func TestMonitoredCacheStatsHitRatio(t *testing.T) {
	s := NewMonitoredCacheStats()
	s.RecordHit()
	s.RecordHit()
	s.RecordHit()
	s.RecordMiss()

	snap := s.Snapshot()
	if snap.Hits != 3 || snap.Misses != 1 {
		t.Errorf("unexpected snapshot: %+v", snap)
	}
	if s.HitRatio() != 0.75 {
		t.Errorf("expected hit ratio 0.75, got %v", s.HitRatio())
	}
}

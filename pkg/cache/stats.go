package cache

import (
	"sync/atomic"
)

type CacheTelemetry struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

type MonitoredCacheStats struct {
	hits      uint64
	misses    uint64
	evictions uint64
}

func NewMonitoredCacheStats() *MonitoredCacheStats {
	return &MonitoredCacheStats{}
}

func (s *MonitoredCacheStats) RecordHit() {
	atomic.AddUint64(&s.hits, 1)
}

func (s *MonitoredCacheStats) RecordMiss() {
	atomic.AddUint64(&s.misses, 1)
}

func (s *MonitoredCacheStats) RecordEviction() {
	atomic.AddUint64(&s.evictions, 1)
}

func (s *MonitoredCacheStats) Snapshot() CacheTelemetry {
	return CacheTelemetry{
		Hits:      atomic.LoadUint64(&s.hits),
		Misses:    atomic.LoadUint64(&s.misses),
		Evictions: atomic.LoadUint64(&s.evictions),
	}
}

func (s *MonitoredCacheStats) HitRatio() float64 {
	hits := atomic.LoadUint64(&s.hits)
	misses := atomic.LoadUint64(&s.misses)
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

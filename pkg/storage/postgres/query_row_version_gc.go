package postgres

import (
	"context"
	"sync"
	"time"
)

// DeadTupleStats records table dead tuple metrics for GC estimation.
type DeadTupleStats struct {
	TableName     string    `json:"table_name"`
	LiveTuples    int64     `json:"live_tuples"`
	DeadTuples    int64     `json:"dead_tuples"`
	DeadRatio     float64   `json:"dead_ratio"`
	NeedsVacuum   bool      `json:"needs_vacuum"`
	EvaluatedAt   time.Time `json:"evaluated_at"`
}

// RowVersionGCEstimator identifies tables with high churn needing autovacuum tuning.
type RowVersionGCEstimator struct {
	mu             sync.RWMutex
	vacuumCutoff   float64
	observedTables map[string]DeadTupleStats
}

// NewRowVersionGCEstimator initializes an estimator with dead tuple ratio threshold.
func NewRowVersionGCEstimator(deadRatioCutoff float64) *RowVersionGCEstimator {
	if deadRatioCutoff <= 0.0 || deadRatioCutoff >= 1.0 {
		deadRatioCutoff = 0.20 // 20% default
	}
	return &RowVersionGCEstimator{
		vacuumCutoff:   deadRatioCutoff,
		observedTables: make(map[string]DeadTupleStats),
	}
}

// RecordStats assesses dead tuple count.
func (e *RowVersionGCEstimator) RecordStats(ctx context.Context, tableName string, live, dead int64) DeadTupleStats {
	e.mu.Lock()
	defer e.mu.Unlock()

	total := live + dead
	ratio := 0.0
	if total > 0 {
		ratio = float64(dead) / float64(total)
	}

	stats := DeadTupleStats{
		TableName:   tableName,
		LiveTuples:  live,
		DeadTuples:  dead,
		DeadRatio:   ratio,
		NeedsVacuum: ratio >= e.vacuumCutoff && dead > 1000,
		EvaluatedAt: time.Now(),
	}

	e.observedTables[tableName] = stats
	return stats
}

// CandidateVacuumTables returns tables requiring vacuuming.
func (e *RowVersionGCEstimator) CandidateVacuumTables() []DeadTupleStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var out []DeadTupleStats
	for _, s := range e.observedTables {
		if s.NeedsVacuum {
			out = append(out, s)
		}
	}
	return out
}

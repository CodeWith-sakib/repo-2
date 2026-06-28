package postgres

import (
	"sync"
	"time"
)

// IndexBloatReport models estimated disk space wastage inside b-tree indexes.
type IndexBloatReport struct {
	IndexName        string    `json:"index_name"`
	TableName        string    `json:"table_name"`
	TotalSizeBytes   int64     `json:"total_size_bytes"`
	BloatBytes       int64     `json:"bloat_bytes"`
	BloatRatio       float64   `json:"bloat_ratio"`
	NeedsReindex     bool      `json:"needs_reindex"`
	AnalyzedAt       time.Time `json:"analyzed_at"`
}

// IndexBloatEstimator calculates B-tree index fragmentation and suggests reindexing schedules.
type IndexBloatEstimator struct {
	mu             sync.RWMutex
	reindexCutoff  float64
	recordedBloat  map[string]IndexBloatReport
}

// NewIndexBloatEstimator initializes an estimator with bloat ratio alert threshold (e.g. 0.40 = 40%).
func NewIndexBloatEstimator(reindexThreshold float64) *IndexBloatEstimator {
	if reindexThreshold <= 0.0 || reindexThreshold >= 1.0 {
		reindexThreshold = 0.35
	}
	return &IndexBloatEstimator{
		reindexCutoff: reindexThreshold,
		recordedBloat: make(map[string]IndexBloatReport),
	}
}

// RecordIndexStats assesses index usage and dead leaf pages.
func (e *IndexBloatEstimator) RecordIndexStats(tableName, indexName string, totalBytes, liveBytes int64) IndexBloatReport {
	e.mu.Lock()
	defer e.mu.Unlock()

	bloatBytes := totalBytes - liveBytes
	if bloatBytes < 0 {
		bloatBytes = 0
	}

	ratio := 0.0
	if totalBytes > 0 {
		ratio = float64(bloatBytes) / float64(totalBytes)
	}

	report := IndexBloatReport{
		IndexName:      indexName,
		TableName:      tableName,
		TotalSizeBytes: totalBytes,
		BloatBytes:     bloatBytes,
		BloatRatio:     ratio,
		NeedsReindex:   ratio >= e.reindexCutoff && totalBytes > 10*1024*1024, // > 10MB
		AnalyzedAt:     time.Now(),
	}

	e.recordedBloat[indexName] = report
	return report
}

// CandidateReindexes returns all indexes marked for REINDEX CONCURRENTLY.
func (e *IndexBloatEstimator) CandidateReindexes() []IndexBloatReport {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var candidates []IndexBloatReport
	for _, rep := range e.recordedBloat {
		if rep.NeedsReindex {
			candidates = append(candidates, rep)
		}
	}
	return candidates
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// QueryPerformanceMetrics contains aggregated latency metrics for a query pattern.
type QueryPerformanceMetrics struct {
	QueryID      string
	Calls        int64
	TotalTimeMs  float64
	MeanTimeMs   float64
	RowsReturned int64
	LastSampled  time.Time
}

// StatsCollector fetches and normalizes execution statistics from PostgreSQL metrics views.
type StatsCollector struct {
	db *sql.DB
}

// NewStatsCollector creates a new query stats collector.
func NewStatsCollector(db *sql.DB) *StatsCollector {
	return &StatsCollector{db: db}
}

// CollectSlowQueries fetches execution metrics or returns simulated telemetry if running against mock DB.
func (sc *StatsCollector) CollectSlowQueries(ctx context.Context, minMeanTimeMs float64, limit int) ([]QueryPerformanceMetrics, error) {
	if limit <= 0 {
		limit = 50
	}

	if sc.db == nil {
		// Return simulated telemetry metrics for mock environments
		return []QueryPerformanceMetrics{
			{
				QueryID:      "q_workflow_runs_by_tenant",
				Calls:        14500,
				TotalTimeMs:  18500.0,
				MeanTimeMs:   1.27,
				RowsReturned: 14500,
				LastSampled:  time.Now().UTC(),
			},
			{
				QueryID:      "q_step_runs_pending_acquire",
				Calls:        92000,
				TotalTimeMs:  48000.0,
				MeanTimeMs:   0.52,
				RowsReturned: 91800,
				LastSampled:  time.Now().UTC(),
			},
		}, nil
	}

	query := `
		SELECT queryid::text, calls, total_exec_time, mean_exec_time, rows
		FROM pg_stat_statements
		WHERE mean_exec_time >= $1
		ORDER BY mean_exec_time DESC
		LIMIT $2
	`

	rows, err := sc.db.QueryContext(ctx, query, minMeanTimeMs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed querying pg_stat_statements: %w", err)
	}
	defer rows.Close()

	var result []QueryPerformanceMetrics
	now := time.Now().UTC()

	for rows.Next() {
		var m QueryPerformanceMetrics
		if err := rows.Scan(&m.QueryID, &m.Calls, &m.TotalTimeMs, &m.MeanTimeMs, &m.RowsReturned); err != nil {
			return nil, fmt.Errorf("failed scanning query metrics: %w", err)
		}
		m.LastSampled = now
		result = append(result, m)
	}

	return result, rows.Err()
}

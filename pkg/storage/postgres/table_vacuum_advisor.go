package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TableBloatAnalysis contains dead tuple analysis and recommended vacuum actions.
type TableBloatAnalysis struct {
	TableName      string
	LiveTuples     int64
	DeadTuples     int64
	DeadRatioPct   float64
	LastAutovacuum time.Time
	RequiresVacuum bool
}

// TableVacuumAdvisor inspects dead tuple ratios and recommends maintenance schedules.
type TableVacuumAdvisor struct {
	db *sql.DB
}

// NewTableVacuumAdvisor creates a vacuum advisor instance.
func NewTableVacuumAdvisor(db *sql.DB) *TableVacuumAdvisor {
	return &TableVacuumAdvisor{db: db}
}

// AnalyzeTableDeadTuples evaluates table bloat against a dead ratio threshold (e.g. 15%).
func (a *TableVacuumAdvisor) AnalyzeTableDeadTuples(ctx context.Context, thresholdPct float64) ([]TableBloatAnalysis, error) {
	if thresholdPct <= 0 {
		thresholdPct = 10.0
	}

	if a.db == nil {
		// Return mock analysis for simulated environments
		return []TableBloatAnalysis{
			{
				TableName:      "workflow_run_steps",
				LiveTuples:     500000,
				DeadTuples:     125000,
				DeadRatioPct:   20.0,
				LastAutovacuum: time.Now().UTC().Add(-48 * time.Hour),
				RequiresVacuum: true,
			},
			{
				TableName:      "workflow_definitions",
				LiveTuples:     1200,
				DeadTuples:     10,
				DeadRatioPct:   0.8,
				LastAutovacuum: time.Now().UTC().Add(-6 * time.Hour),
				RequiresVacuum: false,
			},
		}, nil
	}

	query := `
		SELECT relname, n_live_tup, n_dead_tup,
			COALESCE(last_autovacuum, '1970-01-01'::timestamp)
		FROM pg_stat_user_tables
		ORDER BY n_dead_tup DESC
	`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying pg_stat_user_tables: %w", err)
	}
	defer rows.Close()

	var results []TableBloatAnalysis
	for rows.Next() {
		var (
			name     string
			live     int64
			dead     int64
			lastAuto time.Time
		)
		if err := rows.Scan(&name, &live, &dead, &lastAuto); err != nil {
			return nil, fmt.Errorf("failed scanning table stats: %w", err)
		}

		total := live + dead
		var ratio float64
		if total > 0 {
			ratio = (float64(dead) / float64(total)) * 100.0
		}

		results = append(results, TableBloatAnalysis{
			TableName:      name,
			LiveTuples:     live,
			DeadTuples:     dead,
			DeadRatioPct:   ratio,
			LastAutovacuum: lastAuto,
			RequiresVacuum: ratio >= thresholdPct && dead > 1000,
		})
	}

	return results, rows.Err()
}

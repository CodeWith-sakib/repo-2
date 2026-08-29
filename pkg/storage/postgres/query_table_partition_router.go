package postgres

import (
	"fmt"
	"sync"
	"time"
)

// PartitionRange describes upper and lower timestamp bounds for a time-range partition table.
type PartitionRange struct {
	TableName string    `json:"table_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// TablePartitionRouter generates partition DDL and selects partition tables for time-series workflow events.
type TablePartitionRouter struct {
	mu           sync.RWMutex
	baseTable    string
	intervalDays int
}

// NewTablePartitionRouter creates a partition manager.
func NewTablePartitionRouter(baseTable string, intervalDays int) *TablePartitionRouter {
	if baseTable == "" {
		baseTable = "kf_workflow_events"
	}
	if intervalDays <= 0 {
		intervalDays = 7 // weekly default
	}
	return &TablePartitionRouter{
		baseTable:    baseTable,
		intervalDays: intervalDays,
	}
}

// PartitionForTime returns the target partition table name for a given event timestamp.
func (r *TablePartitionRouter) PartitionForTime(t time.Time) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	year, week := t.ISOWeek()
	return fmt.Sprintf("%s_y%d_w%02d", r.baseTable, year, week)
}

// GenerateCreatePartitionSQL creates DDL for the weekly partition table.
func (r *TablePartitionRouter) GenerateCreatePartitionSQL(t time.Time) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	partName := r.PartitionForTime(t)
	start := t.Truncate(24 * time.Hour)
	end := start.Add(time.Duration(r.intervalDays) * 24 * time.Hour)

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s');",
		partName, r.baseTable, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

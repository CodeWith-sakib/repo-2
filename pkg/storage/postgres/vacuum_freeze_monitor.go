package postgres

import (
	"context"
	"sync"
	"time"
)

// TableFreezeStatus records current transaction ID age and freeze risk.
type TableFreezeStatus struct {
	TableName         string    `json:"table_name"`
	CurrentTXIDAge    int64     `json:"current_txid_age"`
	AutovacuumFreezeMax int64   `json:"autovacuum_freeze_max"`
	RequiresUrgentFreeze bool   `json:"requires_urgent_freeze"`
	LastAnalyzedAt    time.Time `json:"last_analyzed_at"`
}

// VacuumFreezeMonitor detects impending transaction ID wraparound risks in Postgres tables.
type VacuumFreezeMonitor struct {
	mu           sync.RWMutex
	freezeLimit  int64
	warnRatio    float64
	tableRecords map[string]TableFreezeStatus
}

// NewVacuumFreezeMonitor constructs a freeze risk watcher.
func NewVacuumFreezeMonitor(freezeMaxAge int64, warnThresholdRatio float64) *VacuumFreezeMonitor {
	if freezeMaxAge <= 0 {
		freezeMaxAge = 200_000_000 // default autovacuum_freeze_max_age
	}
	if warnThresholdRatio <= 0.0 || warnThresholdRatio >= 1.0 {
		warnThresholdRatio = 0.75
	}
	return &VacuumFreezeMonitor{
		freezeLimit:  freezeMaxAge,
		warnRatio:    warnThresholdRatio,
		tableRecords: make(map[string]TableFreezeStatus),
	}
}

// RecordTableAge evaluates a table's current transaction ID age against wraparound limits.
func (m *VacuumFreezeMonitor) RecordTableAge(ctx context.Context, tableName string, txidAge int64) TableFreezeStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	threshold := int64(float64(m.freezeLimit) * m.warnRatio)
	urgent := txidAge >= threshold

	status := TableFreezeStatus{
		TableName:            tableName,
		CurrentTXIDAge:       txidAge,
		AutovacuumFreezeMax:  m.freezeLimit,
		RequiresUrgentFreeze: urgent,
		LastAnalyzedAt:       time.Now(),
	}
	m.tableRecords[tableName] = status
	return status
}

// GetUrgentTables returns all tables exceeding the wraparound safety margin.
func (m *VacuumFreezeMonitor) GetUrgentTables() []TableFreezeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var urgent []TableFreezeStatus
	for _, s := range m.tableRecords {
		if s.RequiresUrgentFreeze {
			urgent = append(urgent, s)
		}
	}
	return urgent
}

package postgres

import (
	"sync"
	"time"
)

type VacuumTelemetry struct {
	mu           sync.RWMutex
	LastVacuum   time.Time
	TuplesPruned int64
}

func NewVacuumTelemetry() *VacuumTelemetry {
	return &VacuumTelemetry{}
}

func (v *VacuumTelemetry) Record(tuples int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.LastVacuum = time.Now().UTC()
	v.TuplesPruned += tuples
}

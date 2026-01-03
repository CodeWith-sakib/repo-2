package postgres

import (
	"database/sql"
	"math"
	"time"
)

// DynamicPoolProfile represents recommended connection pool settings based on workload.
type DynamicPoolProfile struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ConnectionPoolTuner calculates optimal database/sql connection pool parameters.
type ConnectionPoolTuner struct{}

// NewConnectionPoolTuner creates a new pool tuner.
func NewConnectionPoolTuner() *ConnectionPoolTuner {
	return &ConnectionPoolTuner{}
}

// ComputeProfile derives pool settings from CPU core count and expected peak concurrency.
func (t *ConnectionPoolTuner) ComputeProfile(cpuCores int, peakConcurrency int, isSSD bool) DynamicPoolProfile {
	if cpuCores <= 0 {
		cpuCores = 4
	}
	if peakConcurrency <= 0 {
		peakConcurrency = 20
	}

	// Classic PostgreSQL formula: pool_size = ((core_count * 2) + effective_spindle_count)
	spindleBonus := 1
	if isSSD {
		spindleBonus = 4
	}

	basePool := (cpuCores * 2) + spindleBonus
	maxOpen := int(math.Min(float64(peakConcurrency), float64(basePool*3)))
	if maxOpen < basePool {
		maxOpen = basePool
	}

	maxIdle := maxOpen / 2
	if maxIdle < 2 {
		maxIdle = 2
	}

	return DynamicPoolProfile{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// ApplyConfig configures an active sql.DB instance with the profile.
func (t *ConnectionPoolTuner) ApplyConfig(db *sql.DB, profile DynamicPoolProfile) {
	if db == nil {
		return
	}
	db.SetMaxOpenConns(profile.MaxOpenConns)
	db.SetMaxIdleConns(profile.MaxIdleConns)
	db.SetConnMaxLifetime(profile.ConnMaxLifetime)
	db.SetConnMaxIdleTime(profile.ConnMaxIdleTime)
}

package postgres

import (
	"fmt"
	"sync"
)

// VacuumCostConfig models Postgres vacuum cost delay parameters.
type VacuumCostConfig struct {
	CostLimit int `json:"cost_limit"`
	CostDelayMs int `json:"cost_delay_ms"`
	PageHitCost int `json:"page_hit_cost"`
	PageMissCost int `json:"page_miss_cost"`
	PageDirtyCost int `json:"page_dirty_cost"`
}

// VacuumCostLimiter dynamically tunes autovacuum I/O throttles based on query latency.
type VacuumCostLimiter struct {
	mu     sync.RWMutex
	config VacuumCostConfig
}

// NewVacuumCostLimiter initializes vacuum cost controller.
func NewVacuumCostLimiter(cfg VacuumCostConfig) *VacuumCostLimiter {
	if cfg.CostLimit <= 0 {
		cfg.CostLimit = 200
	}
	if cfg.CostDelayMs <= 0 {
		cfg.CostDelayMs = 2
	}
	return &VacuumCostLimiter{
		config: cfg,
	}
}

// AdjustForHighLoad reduces vacuum aggressive I/O during daytime peaks.
func (l *VacuumCostLimiter) AdjustForHighLoad() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.CostLimit = 100
	l.config.CostDelayMs = 10
	return fmt.Sprintf("ALTER SYSTEM SET autovacuum_vacuum_cost_limit = %d; ALTER SYSTEM SET autovacuum_vacuum_cost_delay = %d;",
		l.config.CostLimit, l.config.CostDelayMs)
}

// AdjustForOffPeak elevates vacuum throughput during nightly maintenance windows.
func (l *VacuumCostLimiter) AdjustForOffPeak() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.CostLimit = 1000
	l.config.CostDelayMs = 0
	return fmt.Sprintf("ALTER SYSTEM SET autovacuum_vacuum_cost_limit = %d; ALTER SYSTEM SET autovacuum_vacuum_cost_delay = %d;",
		l.config.CostLimit, l.config.CostDelayMs)
}

// CurrentConfig returns current vacuum tuning parameters.
func (l *VacuumCostLimiter) CurrentConfig() VacuumCostConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

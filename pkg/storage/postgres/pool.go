package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type PoolStats struct {
	TotalQueries   int64
	ActiveConns    int
	IdleConns      int
	OpenConns      int
	WaitCount      int64
	WaitDuration   time.Duration
	MaxIdleClosed  int64
	MaxLifeClosed  int64
}

type PoolMonitor struct {
	db           *sql.DB
	mu           sync.RWMutex
	queryCount   int64
	lastPing     time.Time
	isHealthy    bool
	stopCh       chan struct{}
}

func NewPoolMonitor(db *sql.DB, probeInterval time.Duration) *PoolMonitor {
	if probeInterval <= 0 {
		probeInterval = 30 * time.Second
	}
	pm := &PoolMonitor{
		db:        db,
		isHealthy: true,
		stopCh:    make(chan struct{}),
	}
	go pm.healthLoop(probeInterval)
	return pm
}

func (pm *PoolMonitor) RecordQuery() {
	atomic.AddInt64(&pm.queryCount, 1)
}

func (pm *PoolMonitor) Stats() PoolStats {
	dbStats := pm.db.Stats()
	return PoolStats{
		TotalQueries:  atomic.LoadInt64(&pm.queryCount),
		ActiveConns:   dbStats.InUse,
		IdleConns:     dbStats.Idle,
		OpenConns:     dbStats.OpenConnections,
		WaitCount:     dbStats.WaitCount,
		WaitDuration:  dbStats.WaitDuration,
		MaxIdleClosed: dbStats.MaxIdleClosed,
		MaxLifeClosed: dbStats.MaxLifetimeClosed,
	}
}

func (pm *PoolMonitor) IsHealthy() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.isHealthy
}

func (pm *PoolMonitor) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := pm.db.PingContext(ctx)
	pm.mu.Lock()
	pm.lastPing = time.Now()
	pm.isHealthy = (err == nil)
	pm.mu.Unlock()

	return err
}

func (pm *PoolMonitor) healthLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = pm.Ping(ctx)
			cancel()
		}
	}
}

func (pm *PoolMonitor) Stop() {
	select {
	case <-pm.stopCh:
	default:
		close(pm.stopCh)
	}
}

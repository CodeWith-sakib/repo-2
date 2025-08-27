package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// StarvationRecord tracks how long a tenant's tasks have been waiting without execution.
type StarvationRecord struct {
	TenantID       string
	EnqueuedAt     time.Time
	LastExecutedAt time.Time
	WaitCount      int64
}

// StarvationGuard monitors tenant task wait times and triggers priority boosts
// when starvation thresholds are exceeded, preventing indefinite denial of service.
type StarvationGuard struct {
	mu             sync.Mutex
	records        map[string]*StarvationRecord
	threshold      time.Duration
	boostFn        func(tenantID string, boost int)
	boostMagnitude int
}

// NewStarvationGuard creates a guard that calls boostFn when a tenant exceeds starvationThreshold.
func NewStarvationGuard(threshold time.Duration, boostMagnitude int, boostFn func(string, int)) *StarvationGuard {
	return &StarvationGuard{
		records:        make(map[string]*StarvationRecord),
		threshold:      threshold,
		boostMagnitude: boostMagnitude,
		boostFn:        boostFn,
	}
}

// RecordEnqueue notes that a tenant has submitted a new task.
func (g *StarvationGuard) RecordEnqueue(tenantID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	rec, exists := g.records[tenantID]
	if !exists {
		rec = &StarvationRecord{
			TenantID:   tenantID,
			EnqueuedAt: time.Now(),
		}
		g.records[tenantID] = rec
	}
	rec.WaitCount++
}

// RecordExecution marks that a tenant's task has been successfully executed.
func (g *StarvationGuard) RecordExecution(tenantID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	rec, exists := g.records[tenantID]
	if !exists {
		return
	}
	rec.LastExecutedAt = time.Now()
	rec.WaitCount = 0
}

// CheckAndBoost scans all known tenants and applies priority boosts for those starving.
// Returns the list of tenants that received a boost.
func (g *StarvationGuard) CheckAndBoost() []string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	var boosted []string

	for tenantID, rec := range g.records {
		if rec.WaitCount == 0 {
			continue
		}

		waitSince := rec.LastExecutedAt
		if waitSince.IsZero() {
			waitSince = rec.EnqueuedAt
		}

		if now.Sub(waitSince) >= g.threshold {
			g.boostFn(tenantID, g.boostMagnitude)
			boosted = append(boosted, tenantID)
		}
	}

	return boosted
}

// Status returns a formatted string report for all tracked tenants.
func (g *StarvationGuard) Status() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.records) == 0 {
		return "starvation_guard: no tenants tracked"
	}

	out := fmt.Sprintf("starvation_guard: %d tenants\n", len(g.records))
	for _, rec := range g.records {
		out += fmt.Sprintf("  tenant=%s waitCount=%d\n", rec.TenantID, rec.WaitCount)
	}
	return out
}

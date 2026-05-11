package postgres

import (
	"sync"
	"time"
)

// CachedPlanInfo stores analyzed query plan estimates and statement execution costs.
type CachedPlanInfo struct {
	QueryFingerprint string
	EstimatedCost    float64
	EstimatedRows    int64
	PlanTreeSummary  string
	CachedAt         time.Time
}

// QueryPlanCache maintains cached query cost plans to avoid repeated EXPLAIN calls.
type QueryPlanCache struct {
	mu    sync.RWMutex
	plans map[string]CachedPlanInfo
	ttl   time.Duration
}

// NewQueryPlanCache creates a query plan cache with entry TTL.
func NewQueryPlanCache(ttl time.Duration) *QueryPlanCache {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &QueryPlanCache{
		plans: make(map[string]CachedPlanInfo),
		ttl:   ttl,
	}
}

// StorePlan caches estimated plan info.
func (c *QueryPlanCache) StorePlan(plan CachedPlanInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	plan.CachedAt = time.Now().UTC()
	c.plans[plan.QueryFingerprint] = plan
}

// GetPlan retrieves cached plan if still valid within TTL.
func (c *QueryPlanCache) GetPlan(fingerprint string, now time.Time) (CachedPlanInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, exists := c.plans[fingerprint]
	if !exists {
		return CachedPlanInfo{}, false
	}

	if now.Sub(info.CachedAt) >= c.ttl {
		return CachedPlanInfo{}, false
	}

	return info, true
}

// Invalidate clears a specific plan or entire cache if empty string.
func (c *QueryPlanCache) Invalidate(fingerprint string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if fingerprint == "" {
		c.plans = make(map[string]CachedPlanInfo)
	} else {
		delete(c.plans, fingerprint)
	}
}

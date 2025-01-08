package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type TenantLimit struct {
	MaxConcurrent int
	RatePerSecond float64
	BurstCapacity int
}

type tenantState struct {
	activeCount int
	tokens      float64
	lastUpdate  time.Time
}

type ConcurrencyGovernor struct {
	mu           sync.Mutex
	limits       map[string]TenantLimit
	tenants      map[string]*tenantState
	defaultLimit TenantLimit
}

func NewConcurrencyGovernor(defaultLimit TenantLimit) *ConcurrencyGovernor {
	return &ConcurrencyGovernor{
		limits:       make(map[string]TenantLimit),
		tenants:      make(map[string]*tenantState),
		defaultLimit: defaultLimit,
	}
}

func (g *ConcurrencyGovernor) SetTenantLimit(tenant string, limit TenantLimit) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.limits[tenant] = limit
}

func (g *ConcurrencyGovernor) Acquire(ctx context.Context, tenant string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	limit, exists := g.limits[tenant]
	if !exists {
		limit = g.defaultLimit
	}

	st, exists := g.tenants[tenant]
	if !exists {
		st = &tenantState{
			tokens:     float64(limit.BurstCapacity),
			lastUpdate: time.Now(),
		}
		g.tenants[tenant] = st
	}

	now := time.Now()
	elapsed := now.Sub(st.lastUpdate).Seconds()
	st.lastUpdate = now
	st.tokens += elapsed * limit.RatePerSecond
	if st.tokens > float64(limit.BurstCapacity) {
		st.tokens = float64(limit.BurstCapacity)
	}

	if limit.MaxConcurrent > 0 && st.activeCount >= limit.MaxConcurrent {
		return fmt.Errorf("%w: tenant %s exceeded max concurrency of %d", core.ErrRateLimited, tenant, limit.MaxConcurrent)
	}

	if st.tokens < 1.0 {
		return fmt.Errorf("%w: tenant %s exceeded rate limit", core.ErrRateLimited, tenant)
	}

	st.tokens -= 1.0
	st.activeCount++
	return nil
}

func (g *ConcurrencyGovernor) Release(tenant string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	st, exists := g.tenants[tenant]
	if exists && st.activeCount > 0 {
		st.activeCount--
	}
}

func (g *ConcurrencyGovernor) GetActive(tenant string) int {
	g.mu.Lock()
	defer g.mu.Unlock()

	st, exists := g.tenants[tenant]
	if !exists {
		return 0
	}
	return st.activeCount
}

package scheduler

import (
	"sync"
)

type TenantQuota struct {
	MaxConcurrentRuns int
	MaxQueuedRuns     int
}

type QuotaEnforcer struct {
	mu           sync.Mutex
	quotas       map[string]TenantQuota
	activeCounts map[string]int
	queuedCounts map[string]int
}

func NewQuotaEnforcer() *QuotaEnforcer {
	return &QuotaEnforcer{
		quotas:       make(map[string]TenantQuota),
		activeCounts: make(map[string]int),
		queuedCounts: make(map[string]int),
	}
}

func (qe *QuotaEnforcer) SetQuota(tenant string, quota TenantQuota) {
	qe.mu.Lock()
	defer qe.mu.Unlock()
	qe.quotas[tenant] = quota
}

func (qe *QuotaEnforcer) CanEnqueue(tenant string) bool {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	quota, exists := qe.quotas[tenant]
	if !exists || quota.MaxQueuedRuns <= 0 {
		return true
	}
	return qe.queuedCounts[tenant] < quota.MaxQueuedRuns
}

func (qe *QuotaEnforcer) Enqueue(tenant string) bool {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	quota, exists := qe.quotas[tenant]
	if exists && quota.MaxQueuedRuns > 0 && qe.queuedCounts[tenant] >= quota.MaxQueuedRuns {
		return false
	}
	qe.queuedCounts[tenant]++
	return true
}

func (qe *QuotaEnforcer) CanStart(tenant string) bool {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	quota, exists := qe.quotas[tenant]
	if !exists || quota.MaxConcurrentRuns <= 0 {
		return true
	}
	return qe.activeCounts[tenant] < quota.MaxConcurrentRuns
}

func (qe *QuotaEnforcer) Start(tenant string) bool {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	quota, exists := qe.quotas[tenant]
	if exists && quota.MaxConcurrentRuns > 0 && qe.activeCounts[tenant] >= quota.MaxConcurrentRuns {
		return false
	}
	if qe.queuedCounts[tenant] > 0 {
		qe.queuedCounts[tenant]--
	}
	qe.activeCounts[tenant]++
	return true
}

func (qe *QuotaEnforcer) Finish(tenant string) {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	if qe.activeCounts[tenant] > 0 {
		qe.activeCounts[tenant]--
	}
}

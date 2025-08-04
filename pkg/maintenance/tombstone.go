package maintenance

import (
	"context"
	"sync"
	"time"
)

type TombstonePurgePolicy struct {
	MaxAge    time.Duration
	BatchSize int
}

type TombstoneVacuum struct {
	mu     sync.Mutex
	policy TombstonePurgePolicy
}

func NewTombstoneVacuum(policy TombstonePurgePolicy) *TombstoneVacuum {
	if policy.MaxAge <= 0 {
		policy.MaxAge = 14 * 24 * time.Hour
	}
	if policy.BatchSize <= 0 {
		policy.BatchSize = 500
	}
	return &TombstoneVacuum{policy: policy}
}

func (tv *TombstoneVacuum) PurgeEligible(deletedAt time.Time) bool {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	cutoff := time.Now().UTC().Add(-tv.policy.MaxAge)
	return deletedAt.Before(cutoff)
}

func (tv *TombstoneVacuum) RunVacuum(ctx context.Context, tombstones []time.Time) int {
	purged := 0
	for _, ts := range tombstones {
		if tv.PurgeEligible(ts) {
			purged++
			if purged >= tv.policy.BatchSize {
				break
			}
		}
	}
	return purged
}

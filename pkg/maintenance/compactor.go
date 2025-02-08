package maintenance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type CompactionPolicy struct {
	RetentionPeriod time.Duration
	BatchSize       int
	DryRun          bool
}

type CompactionStats struct {
	RunsEvaluated int
	RunsPruned    int
	StepsPruned   int
	Duration      time.Duration
}

type HistoryCompactor struct {
	store  storage.RunStore
	policy CompactionPolicy
	mu     sync.Mutex
}

func NewHistoryCompactor(store storage.RunStore, policy CompactionPolicy) *HistoryCompactor {
	if policy.RetentionPeriod <= 0 {
		policy.RetentionPeriod = 30 * 24 * time.Hour // 30 days
	}
	if policy.BatchSize <= 0 {
		policy.BatchSize = 100
	}
	return &HistoryCompactor{
		store:  store,
		policy: policy,
	}
}

func (c *HistoryCompactor) Compact(ctx context.Context) (*CompactionStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	start := time.Now()
	stats := &CompactionStats{}
	cutoff := time.Now().Add(-c.policy.RetentionPeriod)

	runs, _, err := c.store.ListRuns(ctx, storage.RunFilter{
		Pagination: storage.Pagination{Limit: c.policy.BatchSize},
	})
	if err != nil {
		return nil, fmt.Errorf("failed listing runs for compaction: %w", err)
	}

	for _, r := range runs {
		stats.RunsEvaluated++
		if !r.State.IsTerminal() {
			continue
		}
		if r.FinishedAt != nil && r.FinishedAt.Before(cutoff) {
			if !c.policy.DryRun {
				// Delete or archive terminal run
				stats.RunsPruned++
			}
		}
	}

	stats.Duration = time.Since(start)
	return stats, nil
}

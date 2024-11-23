package maintenance

import (
	"context"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Cleaner struct {
	store           storage.EngineStore
	retentionWindow time.Duration
	interval        time.Duration
}

func NewCleaner(store storage.EngineStore, retention, interval time.Duration) *Cleaner {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	if interval <= 0 {
		interval = time.Hour
	}
	return &Cleaner{
		store:           store,
		retentionWindow: retention,
		interval:        interval,
	}
}

func (c *Cleaner) RunOnce(ctx context.Context) (int, error) {
	// Requeue any orphaned leases
	requeued, err := c.store.RequeueOrphaned(ctx)
	if err != nil {
		return 0, err
	}
	return requeued, nil
}

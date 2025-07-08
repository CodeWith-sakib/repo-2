package maintenance

import (
	"sync"
	"time"
)

type OrphanTaskCleaner struct {
	mu           sync.Mutex
	staleThreshold time.Duration
}

func NewOrphanTaskCleaner(threshold time.Duration) *OrphanTaskCleaner {
	if threshold <= 0 {
		threshold = 10 * time.Minute
	}
	return &OrphanTaskCleaner{staleThreshold: threshold}
}

func (c *OrphanTaskCleaner) IsStale(lastHeartbeat time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(lastHeartbeat) > c.staleThreshold
}

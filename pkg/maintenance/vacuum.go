package maintenance

import (
	"context"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type VacuumStats struct {
	OrphanedTasksReclaimed int
	PartitionsScanned      int
	Duration               time.Duration
}

type StorageVacuum struct {
	mu         sync.Mutex
	store      storage.QueueStore
	leaseGrace time.Duration
}

func NewStorageVacuum(store storage.QueueStore, leaseGrace time.Duration) *StorageVacuum {
	if leaseGrace <= 0 {
		leaseGrace = 15 * time.Second
	}
	return &StorageVacuum{
		store:      store,
		leaseGrace: leaseGrace,
	}
}

func (v *StorageVacuum) ReclaimOrphaned(ctx context.Context) (*VacuumStats, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	start := time.Now()
	reclaimed, err := v.store.RequeueOrphaned(ctx)
	if err != nil {
		return nil, err
	}

	return &VacuumStats{
		OrphanedTasksReclaimed: reclaimed,
		PartitionsScanned:      1,
		Duration:               time.Since(start),
	}, nil
}

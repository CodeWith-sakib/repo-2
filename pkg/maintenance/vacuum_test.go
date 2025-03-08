package maintenance

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestStorageVacuum(t *testing.T) {
	store := memory.NewStore()
	vacuum := NewStorageVacuum(store, 10*time.Second)

	stats, err := vacuum.ReclaimOrphaned(context.Background())
	if err != nil {
		t.Fatalf("reclaim orphaned failed: %v", err)
	}

	if stats.PartitionsScanned != 1 {
		t.Errorf("expected 1 partition scanned, got %d", stats.PartitionsScanned)
	}
}

package maintenance

import (
	"context"
	"testing"
	"time"
)

func TestTombstoneVacuum(t *testing.T) {
	vac := NewTombstoneVacuum(TombstonePurgePolicy{
		MaxAge:    1 * time.Hour,
		BatchSize: 10,
	})

	now := time.Now().UTC()
	timestamps := []time.Time{
		now.Add(-2 * time.Hour), // Eligible
		now.Add(-3 * time.Hour), // Eligible
		now.Add(-10 * time.Minute), // Ineligible
	}

	purged := vac.RunVacuum(context.Background(), timestamps)
	if purged != 2 {
		t.Errorf("expected 2 purged tombstones, got %d", purged)
	}
}

package maintenance

import (
	"testing"
	"time"
)

func TestOrphanTaskCleaner(t *testing.T) {
	cleaner := NewOrphanTaskCleaner(50 * time.Millisecond)
	if cleaner.IsStale(time.Now()) {
		t.Error("expected not stale")
	}
	if !cleaner.IsStale(time.Now().Add(-100 * time.Millisecond)) {
		t.Error("expected stale")
	}
}

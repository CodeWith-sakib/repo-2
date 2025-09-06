package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestRateWindow_IncrementAndRate(t *testing.T) {
	w := NewRateWindow(time.Second, 10)
	w.Increment(100)
	w.Increment(50)

	if w.Total() != 150 {
		t.Errorf("expected total 150, got %d", w.Total())
	}

	rate := w.RatePerSecond()
	if rate <= 0 {
		t.Errorf("expected positive rate, got %.2f", rate)
	}

	w.Reset()
	if w.RatePerSecond() != 0 {
		t.Errorf("expected rate 0 after reset")
	}
}

func TestMultiDimensionalCounter_ConcurrentAccess(t *testing.T) {
	mc := NewMultiDimensionalCounter()

	var wg sync.WaitGroup
	labels := []string{"api.ok", "api.err", "db.query", "db.err"}

	for _, label := range labels {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				mc.Inc(l)
			}
		}(label)
	}
	wg.Wait()

	snap := mc.Snapshot()
	for _, l := range labels {
		if snap[l] != 1000 {
			t.Errorf("label %s: expected 1000, got %d", l, snap[l])
		}
	}
}

func TestMultiDimensionalCounter_GetMissingLabel(t *testing.T) {
	mc := NewMultiDimensionalCounter()
	if mc.Get("nonexistent") != 0 {
		t.Error("expected 0 for unknown label")
	}
}

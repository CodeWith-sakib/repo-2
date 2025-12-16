package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadThroughCacheSingleflight(t *testing.T) {
	backend, err := NewSegmentedLRU(50, 50)
	if err != nil {
		t.Fatalf("failed creating SegmentedLRU: %v", err)
	}

	var loadCalls int64
	loader := func(ctx context.Context, key string) (interface{}, error) {
		atomic.AddInt64(&loadCalls, 1)
		time.Sleep(50 * time.Millisecond) // simulate remote database fetch
		return "loaded-" + key, nil
	}

	rtc := NewReadThroughCache(backend, 5*time.Minute, loader)

	var wg sync.WaitGroup
	workers := 10
	results := make([]string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			val, err := rtc.GetOrLoad(context.Background(), "customer:1001")
			if err != nil {
				t.Errorf("worker %d failed: %v", idx, err)
				return
			}
			results[idx] = val.(string)
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt64(&loadCalls) != 1 {
		t.Errorf("expected exactly 1 loader invocation, got %d", atomic.LoadInt64(&loadCalls))
	}

	for _, r := range results {
		if r != "loaded-customer:1001" {
			t.Errorf("unexpected value from read-through: %s", r)
		}
	}
}

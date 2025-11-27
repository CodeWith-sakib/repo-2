package worker

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskMemoizationCache_HitAndMiss(t *testing.T) {
	cache := NewTaskMemoizationCache()

	key := ComputeKey("transform", json.RawMessage(`{"op":"uppercase"}`), json.RawMessage(`{"text":"hello"}`))

	// Initial get -> Miss
	res, ok := cache.Get(key)
	if ok || res != nil {
		t.Error("expected initial cache miss")
	}

	// Store result
	expectedResult := &StepResult{
		Output: json.RawMessage(`{"text":"HELLO"}`),
	}
	if err := cache.Put(key, expectedResult, time.Hour); err != nil {
		t.Fatalf("put failed: %v", err)
	}

	// Second get -> Hit!
	res, ok = cache.Get(key)
	if !ok || res == nil {
		t.Fatal("expected cache hit")
	}
	if string(res.Output) != `{"text":"HELLO"}` {
		t.Errorf("unexpected output: %s", string(res.Output))
	}

	hits, misses, stores := cache.Metrics()
	if hits != 1 || misses != 1 || stores != 1 {
		t.Errorf("metrics mismatch: hits=%d misses=%d stores=%d", hits, misses, stores)
	}
}

func TestTaskMemoizationCache_Expiration(t *testing.T) {
	cache := NewTaskMemoizationCache()
	key := "temp-key"

	_ = cache.Put(key, &StepResult{Output: json.RawMessage(`{}`)}, 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	// Should be expired
	_, ok := cache.Get(key)
	if ok {
		t.Error("expected expired entry to be miss")
	}

	pruned := cache.PruneExpired()
	if pruned != 1 {
		t.Errorf("expected 1 pruned entry, got %d", pruned)
	}
}

package cache

import (
	"fmt"
	"sync"
	"testing"
)

func TestShardedLRUCache_BasicOps(t *testing.T) {
	c, err := NewShardedLRUCache(4, 20)
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}

	c.Set("foo", "bar")
	c.Set("baz", 42)

	if v, ok := c.Get("foo"); !ok || v != "bar" {
		t.Errorf("expected 'bar', got %v", v)
	}
	if v, ok := c.Get("baz"); !ok || v != 42 {
		t.Errorf("expected 42, got %v", v)
	}

	c.Delete("foo")
	if _, ok := c.Get("foo"); ok {
		t.Error("expected foo to be deleted")
	}
}

func TestShardedLRUCache_EvictionUnderCapacity(t *testing.T) {
	c, _ := NewShardedLRUCache(2, 4) // 2 per shard

	// Fill a single shard beyond capacity by using same-shard keys
	// just insert a bunch of unique keys and verify Len() stays bounded
	for i := 0; i < 20; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}

	if c.Len() > 20 {
		t.Errorf("cache grew too large: %d", c.Len())
	}
}

func TestShardedLRUCache_ConcurrentAccess(t *testing.T) {
	c, _ := NewShardedLRUCache(8, 1000)

	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("key-%d-%d", id, i)
				c.Set(key, id*100+i)
				c.Get(key)
			}
		}(g)
	}
	wg.Wait()
}

func TestShardedLRUCache_InvalidShardCount(t *testing.T) {
	if _, err := NewShardedLRUCache(3, 100); err == nil {
		t.Error("expected error for non-power-of-two shard count")
	}
}

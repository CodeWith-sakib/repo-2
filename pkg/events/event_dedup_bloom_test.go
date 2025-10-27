package events

import (
	"fmt"
	"testing"
)

func TestCountingBloomFilter_AddAndContains(t *testing.T) {
	bf, err := NewCountingBloomFilter(1000, 0.01)
	if err != nil {
		t.Fatalf("create bloom filter failed: %v", err)
	}

	// Insert items
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("event-%d", i)
		isNew := bf.Add(key)
		if !isNew {
			t.Errorf("item %s should be new", key)
		}
	}

	// Re-adding item should return isNew = false
	if isNew := bf.Add("event-10"); isNew {
		t.Error("re-added item should not be flagged as new")
	}

	// Verify Contains
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("event-%d", i)
		if !bf.Contains(key) {
			t.Errorf("expected filter to contain %s", key)
		}
	}

	// Non-existent item should not be present
	if bf.Contains("definitely-not-inserted-item-12345") {
		t.Error("filter contained non-existent item")
	}
}

func TestCountingBloomFilter_Remove(t *testing.T) {
	bf, _ := NewCountingBloomFilter(100, 0.05)

	key := "transient-event-42"
	bf.Add(key)

	if !bf.Contains(key) {
		t.Fatal("expected item to be present after Add")
	}

	if !bf.Remove(key) {
		t.Fatal("remove should return true")
	}

	if bf.Contains(key) {
		t.Error("item should not be present after Remove")
	}
}

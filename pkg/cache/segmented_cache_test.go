package cache

import (
	"testing"
)

func TestSegmentedLRU_PromotionAndEviction(t *testing.T) {
	// Probation capacity = 2, Protected capacity = 2
	slru, err := NewSegmentedLRU(2, 2)
	if err != nil {
		t.Fatalf("failed to create SLRU: %v", err)
	}

	// Insert "a" and "b" -> both in probation
	slru.Set("a", 1)
	slru.Set("b", 2)

	if slru.Len() != 2 {
		t.Fatalf("expected 2 items, got %d", slru.Len())
	}

	// Access "a" -> promoted to protected
	val, ok := slru.Get("a")
	if !ok || val != 1 {
		t.Errorf("expected to get a=1, got %v (ok=%v)", val, ok)
	}

	hits, _, promos, _, _ := slru.Metrics()
	if hits != 1 || promos != 1 {
		t.Errorf("expected 1 hit, 1 promo, got hits=%d promos=%d", hits, promos)
	}

	// Insert "c" and "d" -> probation was at capacity (now has "b"), adding "c" fills it, adding "d" evicts "b"
	slru.Set("c", 3)
	slru.Set("d", 4)

	// "b" should have been evicted
	if _, ok := slru.Get("b"); ok {
		t.Error("expected 'b' to be evicted from probation")
	}

	// "a" was protected, should still be there!
	if val, ok := slru.Get("a"); !ok || val != 1 {
		t.Errorf("expected protected item 'a' to persist, got %v (ok=%v)", val, ok)
	}
}

func TestSegmentedLRU_Demotion(t *testing.T) {
	slru, _ := NewSegmentedLRU(1, 2)

	// Put and promote two items to protected
	slru.Set("x", 10)
	slru.Get("x") // promoted to protected

	slru.Set("y", 20)
	slru.Get("y") // promoted to protected (protected full: y, x)

	// Now promote "z"
	slru.Set("z", 30)
	slru.Get("z") // promotes "z", demoting oldest ("x") to probation

	_, _, _, demotions, _ := slru.Metrics()
	if demotions != 1 {
		t.Errorf("expected 1 demotion, got %d", demotions)
	}

	// "x" should now be in probation
	if val, ok := slru.Get("x"); !ok || val != 10 {
		t.Errorf("expected to retrieve demoted item 'x', got %v", val)
	}
}

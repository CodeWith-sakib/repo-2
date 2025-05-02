package cache

import (
	"testing"
)

func TestCountingBloomFilter(t *testing.T) {
	bf := NewCountingBloomFilter(512, 3)

	if bf.MightContain("key-1") {
		t.Error("expected MightContain to return false for empty filter")
	}

	bf.Add("key-1")
	if !bf.MightContain("key-1") {
		t.Error("expected MightContain to return true after Add")
	}

	removed := bf.Remove("key-1")
	if !removed {
		t.Error("expected Remove to succeed")
	}

	if bf.MightContain("key-1") {
		t.Error("expected MightContain to return false after Remove")
	}
}

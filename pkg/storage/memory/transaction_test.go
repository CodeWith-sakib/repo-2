package memory

import (
	"errors"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestMVCCTransactionIsolationAndConflict(t *testing.T) {
	store := NewMVCCStore()

	// Tx1 writes "key1"
	tx1 := store.Begin()
	_ = tx1.Put("key1", "value1")
	if err := tx1.Commit(); err != nil {
		t.Fatalf("tx1 commit failed: %v", err)
	}

	// Tx2 and Tx3 start concurrently after Tx1
	tx2 := store.Begin()
	tx3 := store.Begin()

	// Tx2 reads "key1"
	val, ok, _ := tx2.Get("key1")
	if !ok || val != "value1" {
		t.Fatalf("tx2 expected value1, got %v", val)
	}

	// Tx2 updates "key1" and commits
	_ = tx2.Put("key1", "value2")
	if err := tx2.Commit(); err != nil {
		t.Fatalf("tx2 commit failed: %v", err)
	}

	// Tx3 tries to write "key1" -> conflict because tx2 already committed a newer version
	_ = tx3.Put("key1", "value3")
	err := tx3.Commit()
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict for Tx3, got %v", err)
	}
}

package memory

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestSecondaryIndexOperations(t *testing.T) {
	idx := NewSecondaryIndex()
	id1 := core.NewID("run-1")
	id2 := core.NewID("run-2")

	idx.Index("tenant-a:RUNNING", id1)
	idx.Index("tenant-a:RUNNING", id2)

	res := idx.Lookup("tenant-a:RUNNING")
	if len(res) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(res))
	}

	idx.Unindex("tenant-a:RUNNING", id1)
	res = idx.Lookup("tenant-a:RUNNING")
	if len(res) != 1 || res[0] != id2 {
		t.Fatalf("expected id2 remaining, got %v", res)
	}
}

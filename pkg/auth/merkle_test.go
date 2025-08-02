package auth

import (
	"testing"
)

func TestMerkleTreeRootHash(t *testing.T) {
	blocks := [][]byte{
		[]byte("audit-event-1"),
		[]byte("audit-event-2"),
		[]byte("audit-event-3"),
	}

	tree := NewMerkleTree(blocks)
	root := tree.RootHash()
	if root == "" {
		t.Fatal("expected non-empty root hash")
	}

	tree2 := NewMerkleTree(blocks)
	if tree2.RootHash() != root {
		t.Errorf("expected deterministic root hash")
	}
}

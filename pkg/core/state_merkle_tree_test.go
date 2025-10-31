package core

import (
	"testing"
)

func TestStateMerkleTree_DeterministicRoot(t *testing.T) {
	state1 := map[string]interface{}{
		"step-1": "COMPLETED",
		"step-2": "RUNNING",
		"step-3": "PENDING",
	}

	tree1, err := BuildStateMerkleTree(state1)
	if err != nil {
		t.Fatalf("build tree1 failed: %v", err)
	}

	// Identical state built again should have exact same root
	tree2, err := BuildStateMerkleTree(state1)
	if err != nil {
		t.Fatalf("build tree2 failed: %v", err)
	}

	if tree1.RootHash() != tree2.RootHash() {
		t.Errorf("roots mismatch: %s vs %s", tree1.RootHash(), tree2.RootHash())
	}

	// Mutated state should have completely different root
	stateMutated := map[string]interface{}{
		"step-1": "FAILED", // changed!
		"step-2": "RUNNING",
		"step-3": "PENDING",
	}
	treeMutated, _ := BuildStateMerkleTree(stateMutated)
	if treeMutated.RootHash() == tree1.RootHash() {
		t.Error("mutated state produced identical root hash!")
	}
}

func TestStateMerkleTree_VerifyProof(t *testing.T) {
	state := map[string]interface{}{
		"A": 1,
		"B": 2,
	}
	tree, err := BuildStateMerkleTree(state)
	if err != nil {
		t.Fatalf("build tree failed: %v", err)
	}

	leaf0 := tree.Leaves[0].Hash
	leaf1 := tree.Leaves[1].Hash

	proof := MerkleAuditProof{
		LeafHash: leaf0,
		Path: []ProofElement{
			{Hash: leaf1, IsRight: true},
		},
	}

	if !VerifyProof(tree.RootHash(), proof) {
		t.Error("expected valid proof verification")
	}

	// Corrupted proof should fail
	badProof := MerkleAuditProof{
		LeafHash: "corrupted_hash",
		Path:     proof.Path,
	}
	if VerifyProof(tree.RootHash(), badProof) {
		t.Error("expected corrupted proof to fail verification")
	}
}

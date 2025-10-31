package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// MerkleNode represents an internal or leaf node in the Merkle tree.
type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// MerkleAuditProof contains intermediate sibling hashes proving inclusion of a leaf.
type MerkleAuditProof struct {
	LeafHash string
	Path     []ProofElement
}

// ProofElement contains a sibling hash and direction.
type ProofElement struct {
	Hash    string
	IsRight bool
}

// StateMerkleTree builds a cryptographic audit proof tree from workflow step states.
type StateMerkleTree struct {
	Root   *MerkleNode
	Leaves []*MerkleNode
}

// BuildStateMerkleTree creates a Merkle tree from step state key/value pairs.
func BuildStateMerkleTree(state map[string]interface{}) (*StateMerkleTree, error) {
	if len(state) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return &StateMerkleTree{
			Root: &MerkleNode{Hash: hex.EncodeToString(emptyHash[:])},
		}, nil
	}

	// Sort keys for deterministic leaf ordering
	var keys []string
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var leaves []*MerkleNode
	for _, k := range keys {
		leafData := fmt.Sprintf("%s:%v", k, state[k])
		h := sha256.Sum256([]byte(leafData))
		leaves = append(leaves, &MerkleNode{
			Hash: hex.EncodeToString(h[:]),
		})
	}

	root := buildTree(leaves)
	return &StateMerkleTree{
		Root:   root,
		Leaves: leaves,
	}, nil
}

func buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}

	var nextLevel []*MerkleNode
	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		right := left
		if i+1 < len(nodes) {
			right = nodes[i+1]
		}

		combined := left.Hash + right.Hash
		h := sha256.Sum256([]byte(combined))
		parent := &MerkleNode{
			Hash:  hex.EncodeToString(h[:]),
			Left:  left,
			Right: right,
		}
		nextLevel = append(nextLevel, parent)
	}

	return buildTree(nextLevel)
}

// RootHash returns the hex-encoded Merkle root hash.
func (t *StateMerkleTree) RootHash() string {
	if t.Root == nil {
		return ""
	}
	return t.Root.Hash
}

// VerifyProof verifies that leafHash is included in the tree with expectedRootHash.
func VerifyProof(expectedRootHash string, proof MerkleAuditProof) bool {
	curr := proof.LeafHash
	for _, elem := range proof.Path {
		var combined string
		if elem.IsRight {
			combined = curr + elem.Hash
		} else {
			combined = elem.Hash + curr
		}
		h := sha256.Sum256([]byte(combined))
		curr = hex.EncodeToString(h[:])
	}
	return curr == expectedRootHash
}

package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

type MerkleTree struct {
	Root *MerkleNode
}

func NewMerkleTree(dataBlocks [][]byte) *MerkleTree {
	if len(dataBlocks) == 0 {
		return &MerkleTree{}
	}

	var nodes []*MerkleNode
	for _, block := range dataBlocks {
		h := sha256.Sum256(block)
		nodes = append(nodes, &MerkleNode{Hash: hex.EncodeToString(h[:])})
	}

	for len(nodes) > 1 {
		var level []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				combined := nodes[i].Hash + nodes[i+1].Hash
				h := sha256.Sum256([]byte(combined))
				parent := &MerkleNode{
					Hash:  hex.EncodeToString(h[:]),
					Left:  nodes[i],
					Right: nodes[i+1],
				}
				level = append(level, parent)
			} else {
				level = append(level, nodes[i])
			}
		}
		nodes = level
	}

	return &MerkleTree{Root: nodes[0]}
}

func (m *MerkleTree) RootHash() string {
	if m.Root == nil {
		return ""
	}
	return m.Root.Hash
}

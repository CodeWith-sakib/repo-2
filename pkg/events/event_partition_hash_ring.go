package events

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// PartitionHashRing provides consistent hashing of event keys across physical cluster partitions.
type PartitionHashRing struct {
	mu             sync.RWMutex
	vnodesPerNode  int
	ring           []uint32
	ringMap        map[uint32]string
	nodes          map[string]bool
}

// NewPartitionHashRing creates a consistent hash ring.
func NewPartitionHashRing(vnodesPerNode int) *PartitionHashRing {
	if vnodesPerNode <= 0 {
		vnodesPerNode = 50
	}
	return &PartitionHashRing{
		vnodesPerNode: vnodesPerNode,
		ring:          make([]uint32, 0),
		ringMap:       make(map[uint32]string),
		nodes:         make(map[string]bool),
	}
}

// AddNode adds a partition/node to the ring.
func (r *PartitionHashRing) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodes[node] {
		return
	}
	r.nodes[node] = true

	for i := 0; i < r.vnodesPerNode; i++ {
		h := r.hash(node + "#" + strconv.Itoa(i))
		r.ring = append(r.ring, h)
		r.ringMap[h] = node
	}
	sort.Slice(r.ring, func(i, j int) bool { return r.ring[i] < r.ring[j] })
}

// RemoveNode removes a node and its virtual positions.
func (r *PartitionHashRing) RemoveNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.nodes[node] {
		return
	}
	delete(r.nodes, node)

	newRing := make([]uint32, 0, len(r.ring))
	for _, h := range r.ring {
		if r.ringMap[h] != node {
			newRing = append(newRing, h)
		} else {
			delete(r.ringMap, h)
		}
	}
	r.ring = newRing
}

// GetNode routes an event key to the responsible physical node.
func (r *PartitionHashRing) GetNode(key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ring) == 0 {
		return "", fmt.Errorf("hash ring is empty")
	}

	h := r.hash(key)
	idx := sort.Search(len(r.ring), func(i int) bool {
		return r.ring[i] >= h
	})

	if idx == len(r.ring) {
		idx = 0
	}

	return r.ringMap[r.ring[idx]], nil
}

func (r *PartitionHashRing) hash(val string) uint32 {
	hasher := md5.New()
	hasher.Write([]byte(val))
	sum := hasher.Sum(nil)
	return binary.BigEndian.Uint32(sum[:4])
}

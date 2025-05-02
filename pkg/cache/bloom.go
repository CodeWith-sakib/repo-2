package cache

import (
	"hash/fnv"
	"sync"
)

type CountingBloomFilter struct {
	mu      sync.RWMutex
	size    uint32
	buckets []uint8
	hashes  int
}

func NewCountingBloomFilter(size uint32, hashes int) *CountingBloomFilter {
	if size == 0 {
		size = 1024
	}
	if hashes <= 0 {
		hashes = 3
	}
	return &CountingBloomFilter{
		size:    size,
		buckets: make([]uint8, size),
		hashes:  hashes,
	}
}

func (bf *CountingBloomFilter) hash(item string, seed int) uint32 {
	h := fnv.New32a()
	h.Write([]byte(item))
	h.Write([]byte{byte(seed), byte(seed >> 8)})
	return h.Sum32() % bf.size
}

func (bf *CountingBloomFilter) Add(item string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := 0; i < bf.hashes; i++ {
		idx := bf.hash(item, i)
		if bf.buckets[idx] < 255 {
			bf.buckets[idx]++
		}
	}
}

func (bf *CountingBloomFilter) MightContain(item string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for i := 0; i < bf.hashes; i++ {
		idx := bf.hash(item, i)
		if bf.buckets[idx] == 0 {
			return false
		}
	}
	return true
}

func (bf *CountingBloomFilter) Remove(item string) bool {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if !bf.mightContainLocked(item) {
		return false
	}

	for i := 0; i < bf.hashes; i++ {
		idx := bf.hash(item, i)
		if bf.buckets[idx] > 0 {
			bf.buckets[idx]--
		}
	}
	return true
}

func (bf *CountingBloomFilter) mightContainLocked(item string) bool {
	for i := 0; i < bf.hashes; i++ {
		idx := bf.hash(item, i)
		if bf.buckets[idx] == 0 {
			return false
		}
	}
	return true
}

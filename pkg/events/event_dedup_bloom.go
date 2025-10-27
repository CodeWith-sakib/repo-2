package events

import (
	"fmt"
	"hash/fnv"
	"math"
	"sync"
)

// CountingBloomFilter provides space-efficient set membership with deletion support.
type CountingBloomFilter struct {
	mu        sync.RWMutex
	size      uint64 // number of counter buckets
	numHashes int    // k hash functions
	counters  []uint8
	count     int64
}

// NewCountingBloomFilter creates a filter sized for expectedItems and desired false positive probability.
func NewCountingBloomFilter(expectedItems int, falsePositiveRate float64) (*CountingBloomFilter, error) {
	if expectedItems <= 0 || falsePositiveRate <= 0 || falsePositiveRate >= 1.0 {
		return nil, fmt.Errorf("invalid parameters: items=%d, rate=%.4f", expectedItems, falsePositiveRate)
	}

	// Optimal size m = - (n * ln(p)) / (ln(2)^2)
	n := float64(expectedItems)
	m := math.Ceil(-1.0 * (n * math.Log(falsePositiveRate)) / (math.Pow(math.Log(2), 2)))

	// Optimal hash count k = (m/n) * ln(2)
	k := math.Round((m / n) * math.Log(2))
	if k < 1 {
		k = 1
	}

	size := uint64(m)
	return &CountingBloomFilter{
		size:      size,
		numHashes: int(k),
		counters:  make([]uint8, size),
	}, nil
}

// Add inserts an item into the filter. Returns true if the item was likely not present before.
func (f *CountingBloomFilter) Add(item string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	alreadyPresent := true
	indices := f.getIndices(item)

	for _, idx := range indices {
		if f.counters[idx] == 0 {
			alreadyPresent = false
		}
		if f.counters[idx] < 255 {
			f.counters[idx]++
		}
	}

	f.count++
	return !alreadyPresent
}

// Contains checks if an item might be in the filter.
func (f *CountingBloomFilter) Contains(item string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	indices := f.getIndices(item)
	for _, idx := range indices {
		if f.counters[idx] == 0 {
			return false
		}
	}
	return true
}

// Remove decrements the counters for an item.
func (f *CountingBloomFilter) Remove(item string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	indices := f.getIndices(item)

	// Check if all counters are non-zero before decrementing
	for _, idx := range indices {
		if f.counters[idx] == 0 {
			return false
		}
	}

	for _, idx := range indices {
		if f.counters[idx] > 0 {
			f.counters[idx]--
		}
	}

	f.count--
	return true
}

// EstimatedFalsePositiveRate returns theoretical FPR given current item count.
func (f *CountingBloomFilter) EstimatedFalsePositiveRate() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.size == 0 || f.count == 0 {
		return 0.0
	}

	k := float64(f.numHashes)
	m := float64(f.size)
	n := float64(f.count)

	return math.Pow(1.0-math.Exp(-k*n/m), k)
}

// getIndices computes k bucket indices using Kirsch-Mitzenmacher double hashing.
func (f *CountingBloomFilter) getIndices(item string) []uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(item))
	hash1 := h.Sum64()

	// Secondary hash using DJB2 style constant
	hash2 := uint64(5381)
	for i := 0; i < len(item); i++ {
		hash2 = ((hash2 << 5) + hash2) + uint64(item[i])
	}
	if hash2 == 0 {
		hash2 = 1
	}

	indices := make([]uint64, f.numHashes)
	for i := 0; i < f.numHashes; i++ {
		combined := hash1 + uint64(i)*hash2
		indices[i] = combined % f.size
	}
	return indices
}

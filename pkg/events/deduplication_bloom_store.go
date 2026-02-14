package events

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"time"
)

// DeduplicationBloomStore provides space-efficient event ID deduplication with rolling rotation.
type DeduplicationBloomStore struct {
	mu        sync.RWMutex
	bits      []uint64
	numBits   uint64
	numHashes int
	createdAt time.Time
	ttl       time.Duration
}

// NewDeduplicationBloomStore creates a bloom filter for event deduplication.
func NewDeduplicationBloomStore(expectedEvents int, ttl time.Duration) *DeduplicationBloomStore {
	if expectedEvents <= 0 {
		expectedEvents = 100000
	}
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}

	// m = - (n * ln(p)) / (ln(2)^2), assuming p = 0.01 -> ~9.6 bits per item
	numBits := uint64(expectedEvents * 10)
	words := (numBits + 63) / 64

	return &DeduplicationBloomStore{
		bits:      make([]uint64, words),
		numBits:   numBits,
		numHashes: 4,
		createdAt: time.Now().UTC(),
		ttl:       ttl,
	}
}

// AddOrCheck checks if eventID was already observed; if not, records it.
// Returns true if duplicate, false if freshly added.
func (s *DeduplicationBloomStore) AddOrCheck(eventID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	hashes := s.getHashes(eventID)
	allPresent := true

	for _, h := range hashes {
		idx := h % s.numBits
		wordIdx := idx / 64
		bitIdx := idx % 64
		if (s.bits[wordIdx] & (1 << bitIdx)) == 0 {
			allPresent = false
			s.bits[wordIdx] |= (1 << bitIdx)
		}
	}

	return allPresent
}

// IsExpired checks if bloom filter has exceeded its rotation TTL.
func (s *DeduplicationBloomStore) IsExpired(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return now.Sub(s.createdAt) >= s.ttl
}

func (s *DeduplicationBloomStore) getHashes(key string) []uint64 {
	h := sha256.Sum256([]byte(key))
	h1 := binary.BigEndian.Uint64(h[0:8])
	h2 := binary.BigEndian.Uint64(h[8:16])

	results := make([]uint64, s.numHashes)
	for i := 0; i < s.numHashes; i++ {
		results[i] = h1 + uint64(i)*h2
	}
	return results
}

package worker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// KeyedConcurrencySemaphore bounds concurrent execution slots per tenant or step key.
type KeyedConcurrencySemaphore struct {
	mu       sync.Mutex
	limits   map[string]int
	current  map[string]int
	defaultLimit int
}

// NewKeyedConcurrencySemaphore creates a keyed concurrency semaphore.
func NewKeyedConcurrencySemaphore(defaultLimit int) *KeyedConcurrencySemaphore {
	if defaultLimit <= 0 {
		defaultLimit = 5
	}
	return &KeyedConcurrencySemaphore{
		limits:       make(map[string]int),
		current:      make(map[string]int),
		defaultLimit: defaultLimit,
	}
}

// SetKeyLimit explicitly overrides maximum concurrent slots for a key.
func (s *KeyedConcurrencySemaphore) SetKeyLimit(key string, limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits[key] = limit
}

// TryAcquire attempts non-blocking slot acquisition.
func (s *KeyedConcurrencySemaphore) TryAcquire(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	limit := s.defaultLimit
	if l, exists := s.limits[key]; exists {
		limit = l
	}

	curr := s.current[key]
	if curr < limit {
		s.current[key] = curr + 1
		return true
	}

	return false
}

// Acquire waits up to timeout for an available slot on key.
func (s *KeyedConcurrencySemaphore) Acquire(ctx context.Context, key string, timeout time.Duration) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if s.TryAcquire(key) {
			return nil
		}

		select {
		case <-ctxTimeout.Done():
			return errors.New("timeout acquiring keyed concurrency slot")
		case <-ticker.C:
		}
	}
}

// Release yields an occupied slot for the given key.
func (s *KeyedConcurrencySemaphore) Release(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if curr := s.current[key]; curr > 0 {
		s.current[key] = curr - 1
		if s.current[key] == 0 {
			delete(s.current, key)
		}
	}
}

// ActiveCount returns currently allocated slots for a key.
func (s *KeyedConcurrencySemaphore) ActiveCount(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current[key]
}

package auth

import (
	"sync"
	"time"
)

type TokenBlacklist struct {
	mu     sync.RWMutex
	revoked map[string]time.Time
}

func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		revoked: make(map[string]time.Time),
	}
}

func (b *TokenBlacklist) Revoke(tokenID string, expiresAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.revoked[tokenID] = expiresAt
}

func (b *TokenBlacklist) IsRevoked(tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	exp, exists := b.revoked[tokenID]
	if !exists {
		return false
	}
	return time.Now().Before(exp)
}

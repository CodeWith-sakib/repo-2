package auth

import (
	"errors"
	"sync"
	"time"
)

// TokenRevocationRecord captures revoked token identifier and revocation time.
type TokenRevocationRecord struct {
	TokenID   string
	RevokedAt time.Time
	ExpiresAt time.Time
	Reason    string
}

// OAuth2TokenRevoker maintains a blacklist of revoked bearer tokens and refresh tokens.
type OAuth2TokenRevoker struct {
	mu      sync.RWMutex
	revoked map[string]TokenRevocationRecord
}

// NewOAuth2TokenRevoker creates a token revocation blacklist.
func NewOAuth2TokenRevoker() *OAuth2TokenRevoker {
	return &OAuth2TokenRevoker{
		revoked: make(map[string]TokenRevocationRecord),
	}
}

// Revoke records a token ID as revoked until its original expiration date.
func (r *OAuth2TokenRevoker) Revoke(tokenID string, expiresAt time.Time, reason string) error {
	if tokenID == "" {
		return errors.New("token ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.revoked[tokenID] = TokenRevocationRecord{
		TokenID:   tokenID,
		RevokedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	return nil
}

// IsRevoked checks whether a given token is present in the active revocation list.
func (r *OAuth2TokenRevoker) IsRevoked(tokenID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.revoked[tokenID]
	if !exists {
		return false
	}

	// If the token would have naturally expired anyway, we still consider it invalid
	return true
}

// SweepExpired purges tokens whose natural expiration time has already passed to reclaim memory.
func (r *OAuth2TokenRevoker) SweepExpired(now time.Time) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for id, rec := range r.revoked {
		if !rec.ExpiresAt.IsZero() && rec.ExpiresAt.Before(now) {
			delete(r.revoked, id)
			count++
		}
	}
	return count
}

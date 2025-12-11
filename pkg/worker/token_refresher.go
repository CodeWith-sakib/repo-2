package worker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// AuthToken represents an authenticated identity credential with expiration.
type AuthToken struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}

// IsExpired checks whether token is already expired or within proactive refresh buffer.
func (t *AuthToken) IsExpired(buffer time.Duration) bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	return time.Now().UTC().Add(buffer).After(t.ExpiresAt)
}

// TokenRefreshFunc is called to generate or fetch fresh credentials.
type TokenRefreshFunc func(ctx context.Context) (*AuthToken, error)

// ManagedTokenRefresher coordinates proactive token refreshing across concurrent workers.
type ManagedTokenRefresher struct {
	mu            sync.RWMutex
	currentToken  *AuthToken
	refreshFn     TokenRefreshFunc
	refreshBuffer time.Duration
}

// NewManagedTokenRefresher creates a refresher with specified refresh lead-time buffer.
func NewManagedTokenRefresher(refreshBuffer time.Duration, fn TokenRefreshFunc) *ManagedTokenRefresher {
	if refreshBuffer <= 0 {
		refreshBuffer = 1 * time.Minute
	}
	return &ManagedTokenRefresher{
		refreshBuffer: refreshBuffer,
		refreshFn:     fn,
	}
}

// GetToken returns a valid token, triggering an in-flight refresh if nearing expiration.
func (mtr *ManagedTokenRefresher) GetToken(ctx context.Context) (*AuthToken, error) {
	mtr.mu.RLock()
	tok := mtr.currentToken
	if tok != nil && !tok.IsExpired(mtr.refreshBuffer) {
		mtr.mu.RUnlock()
		return tok, nil
	}
	mtr.mu.RUnlock()

	mtr.mu.Lock()
	defer mtr.mu.Unlock()

	// Double check under write lock
	if mtr.currentToken != nil && !mtr.currentToken.IsExpired(mtr.refreshBuffer) {
		return mtr.currentToken, nil
	}

	if mtr.refreshFn == nil {
		return nil, errors.New("no token refresh function configured")
	}

	fresh, err := mtr.refreshFn(ctx)
	if err != nil {
		return nil, err
	}
	mtr.currentToken = fresh
	return fresh, nil
}

// SetToken explicitly seeds or overrides current active token.
func (mtr *ManagedTokenRefresher) SetToken(token *AuthToken) {
	mtr.mu.Lock()
	defer mtr.mu.Unlock()
	mtr.currentToken = token
}

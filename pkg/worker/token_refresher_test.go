package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagedTokenRefresher(t *testing.T) {
	var callCount int64
	refresher := NewManagedTokenRefresher(100*time.Millisecond, func(ctx context.Context) (*AuthToken, error) {
		atomic.AddInt64(&callCount, 1)
		return &AuthToken{
			AccessToken: "test-bearer-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().UTC().Add(500 * time.Millisecond),
		}, nil
	})

	ctx := context.Background()
	tok1, err := refresher.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error fetching token: %v", err)
	}
	if tok1.AccessToken != "test-bearer-token" {
		t.Errorf("unexpected access token: %s", tok1.AccessToken)
	}

	// Repeated immediate fetch should reuse cached token without invocation
	tok2, err := refresher.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error on second fetch: %v", err)
	}
	if tok2 != tok1 {
		t.Errorf("expected identical token pointer reuse")
	}
	if atomic.LoadInt64(&callCount) != 1 {
		t.Errorf("expected 1 refresh call, got %d", atomic.LoadInt64(&callCount))
	}

	// Test explicit expiration check
	expiredToken := &AuthToken{
		AccessToken: "old",
		ExpiresAt:   time.Now().UTC().Add(-10 * time.Second),
	}
	if !expiredToken.IsExpired(time.Second) {
		t.Error("expected expiredToken to report expired")
	}
}

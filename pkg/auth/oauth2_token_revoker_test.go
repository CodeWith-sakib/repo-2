package auth

import (
	"testing"
	"time"
)

func TestOAuth2TokenRevoker(t *testing.T) {
	revoker := NewOAuth2TokenRevoker()

	exp := time.Now().UTC().Add(1 * time.Hour)
	err := revoker.Revoke("tok_abc_123", exp, "user logout")
	if err != nil {
		t.Fatalf("unexpected error revoking token: %v", err)
	}

	if !revoker.IsRevoked("tok_abc_123") {
		t.Error("expected tok_abc_123 to be revoked")
	}
	if revoker.IsRevoked("tok_other_456") {
		t.Error("expected tok_other_456 to NOT be revoked")
	}

	// Purge expired tokens
	pastExp := time.Now().UTC().Add(-10 * time.Minute)
	_ = revoker.Revoke("tok_past_expired", pastExp, "session expired")

	purged := revoker.SweepExpired(time.Now().UTC())
	if purged != 1 {
		t.Errorf("expected 1 purged token, got %d", purged)
	}
	if revoker.IsRevoked("tok_past_expired") {
		t.Error("expected past expired token to be swept")
	}
}

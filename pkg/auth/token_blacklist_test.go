package auth

import (
	"testing"
	"time"
)

func TestTokenBlacklist(t *testing.T) {
	b := NewTokenBlacklist()
	if b.IsRevoked("tok-1") {
		t.Error("expected not revoked")
	}
	b.Revoke("tok-1", time.Now().Add(1*time.Hour))
	if !b.IsRevoked("tok-1") {
		t.Error("expected revoked")
	}
}

package auth

import (
	"testing"
)

func TestAPIKeyHasher(t *testing.T) {
	hasher := NewAPIKeyHasher("super-secret-salt")

	rawKey, hashed, err := hasher.GenerateKey("kf_test_")
	if err != nil {
		t.Fatalf("unexpected error generating key: %v", err)
	}

	if err := hasher.Verify(rawKey, hashed); err != nil {
		t.Fatalf("expected valid key verification to pass: %v", err)
	}

	// Tampered key
	if err := hasher.Verify(rawKey+"tampered", hashed); err == nil {
		t.Error("expected tampered key to fail verification")
	}

	// Empty inputs
	if err := hasher.Verify("", hashed); err == nil {
		t.Error("expected empty key to fail verification")
	}
}

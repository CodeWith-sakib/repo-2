package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
)

// APIKeyHasher generates, salts, and constant-time verifies high-entropy API authentication keys.
type APIKeyHasher struct {
	pepper []byte
}

// NewAPIKeyHasher creates an API key hasher with server-side pepper secret.
func NewAPIKeyHasher(pepper string) *APIKeyHasher {
	if pepper == "" {
		pepper = "kestrelflow-default-secret-pepper"
	}
	return &APIKeyHasher{
		pepper: []byte(pepper),
	}
}

// GenerateKey produces a cryptographic random key (e.g. kf_live_32hex) and its SHA-256 HMAC hash.
func (h *APIKeyHasher) GenerateKey(prefix string) (rawKey string, hashedKey string, err error) {
	if prefix == "" {
		prefix = "kf_live_"
	}

	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	rawKey = prefix + hex.EncodeToString(b)
	hashedKey = h.Hash(rawKey)
	return rawKey, hashedKey, nil
}

// Hash generates an HMAC-SHA256 hash using the server pepper.
func (h *APIKeyHasher) Hash(rawKey string) string {
	mac := hmac.New(sha256.New, h.pepper)
	mac.Write([]byte(strings.TrimSpace(rawKey)))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify uses constant-time comparison to prevent timing side-channel attacks.
func (h *APIKeyHasher) Verify(rawKey, expectedHash string) error {
	if rawKey == "" || expectedHash == "" {
		return errors.New("rawKey and expectedHash cannot be empty")
	}

	actualHash := h.Hash(rawKey)
	if subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1 {
		return nil
	}

	return errors.New("invalid API key")
}

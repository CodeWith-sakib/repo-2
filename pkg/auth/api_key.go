package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// APIKeyMetadata describes a stored API key entry.
type APIKeyMetadata struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	KeyPrefix    string    `json:"key_prefix"`    // first 8 chars for identification
	HashedSecret string    `json:"hashed_secret"` // SHA-256 hash in hex
	Scopes       []string  `json:"scopes"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Revoked      bool      `json:"revoked"`
}

// GeneratedAPIKey holds the plaintext key returned only once upon creation.
type GeneratedAPIKey struct {
	PlaintextKey string
	Metadata     APIKeyMetadata
}

// APIKeyManager provides generation, hashing, verification, and scope validation of API keys.
type APIKeyManager struct {
	mu   sync.RWMutex
	keys map[string]*APIKeyMetadata // key_prefix -> metadata
}

// NewAPIKeyManager creates a new API key manager instance.
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]*APIKeyMetadata),
	}
}

// Generate creates a new API key with the given scopes and expiration.
func (m *APIKeyManager) Generate(tenantID, name, env string, scopes []string, ttl time.Duration) (*GeneratedAPIKey, error) {
	if env != "live" && env != "test" {
		env = "live"
	}

	// 24 random bytes for high entropy secret
	secretBytes := make([]byte, 24)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random secret: %w", err)
	}
	secretHex := hex.EncodeToString(secretBytes)

	// Format: kf_<env>_<prefix8>_<secret32>
	prefixBytes := make([]byte, 4)
	if _, err := rand.Read(prefixBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random prefix: %w", err)
	}
	prefixHex := hex.EncodeToString(prefixBytes)

	fullKey := fmt.Sprintf("kf_%s_%s_%s", env, prefixHex, secretHex)
	hash := sha256.Sum256([]byte(fullKey))
	hashedHex := hex.EncodeToString(hash[:])

	now := time.Now().UTC()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	meta := APIKeyMetadata{
		ID:           fmt.Sprintf("key-%s", prefixHex),
		TenantID:     tenantID,
		Name:         name,
		KeyPrefix:    prefixHex,
		HashedSecret: hashedHex,
		Scopes:       scopes,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		Revoked:      false,
	}

	m.mu.Lock()
	m.keys[prefixHex] = &meta
	m.mu.Unlock()

	return &GeneratedAPIKey{
		PlaintextKey: fullKey,
		Metadata:     meta,
	}, nil
}

// Validate verifies a plaintext key, checks expiration/revocation, and returns metadata.
func (m *APIKeyManager) Validate(plaintextKey string) (*APIKeyMetadata, error) {
	parts := strings.Split(plaintextKey, "_")
	if len(parts) != 4 || parts[0] != "kf" {
		return nil, fmt.Errorf("invalid API key format")
	}

	prefixHex := parts[2]

	m.mu.RLock()
	meta, exists := m.keys[prefixHex]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("api key not found")
	}

	if meta.Revoked {
		return nil, fmt.Errorf("api key is revoked")
	}

	if !meta.ExpiresAt.IsZero() && time.Now().UTC().After(meta.ExpiresAt) {
		return nil, fmt.Errorf("api key has expired")
	}

	// Constant-time hash comparison
	candidateHash := sha256.Sum256([]byte(plaintextKey))
	candidateHex := hex.EncodeToString(candidateHash[:])

	if subtle.ConstantTimeCompare([]byte(meta.HashedSecret), []byte(candidateHex)) != 1 {
		return nil, fmt.Errorf("invalid api key secret")
	}

	return meta, nil
}

// Revoke invalidates an API key by prefix.
func (m *APIKeyManager) Revoke(prefixHex string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if meta, ok := m.keys[prefixHex]; ok {
		meta.Revoked = true
		return true
	}
	return false
}

// HasScope checks if metadata grants a required scope (supports wildcard "*").
func (meta *APIKeyMetadata) HasScope(required string) bool {
	for _, s := range meta.Scopes {
		if s == "*" || s == required {
			return true
		}
		// Prefix wildcard check: e.g. "workflows:*" covers "workflows:read"
		if strings.HasSuffix(s, ":*") {
			prefix := strings.TrimSuffix(s, ":*")
			if strings.HasPrefix(required, prefix+":") {
				return true
			}
		}
	}
	return false
}

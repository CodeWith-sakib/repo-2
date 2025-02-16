package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type APIKeyRecord struct {
	ID        string    `json:"id"`
	KeyHash   string    `json:"key_hash"`
	TenantID  string    `json:"tenant_id"`
	Role      Role      `json:"role"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

type APIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*APIKeyRecord // keyHash -> record
}

func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{
		keys: make(map[string]*APIKeyRecord),
	}
}

func (s *APIKeyStore) CreateKey(id string, tenantID string, role Role, scopes []string, ttl time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", err
	}
	rawKey := fmt.Sprintf("kf_%s", hex.EncodeToString(rawBytes))
	hash := hashKey(rawKey)

	now := time.Now().UTC()
	rec := &APIKeyRecord{
		ID:        id,
		KeyHash:   hash,
		TenantID:  tenantID,
		Role:      role,
		Scopes:    scopes,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
	}

	s.keys[hash] = rec
	return rawKey, nil
}

func (s *APIKeyStore) Authenticate(rawKey string) (*APIKeyRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash := hashKey(rawKey)
	rec, exists := s.keys[hash]
	if !exists || rec.Revoked {
		return nil, fmt.Errorf("invalid or revoked api key")
	}

	if !rec.ExpiresAt.IsZero() && time.Now().UTC().After(rec.ExpiresAt) {
		return nil, fmt.Errorf("api key has expired")
	}

	return rec, nil
}

func (s *APIKeyStore) Revoke(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.keys {
		if rec.ID == id {
			rec.Revoked = true
			return true
		}
	}
	return false
}

func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

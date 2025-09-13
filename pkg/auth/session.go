package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SessionState is the lifecycle state of a user session.
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStateExpired   SessionState = "expired"
	SessionStateRevoked   SessionState = "revoked"
	SessionStateSuspended SessionState = "suspended"
)

// Session represents an authenticated user's active session with metadata.
type Session struct {
	ID         string
	UserID     string
	TenantID   string
	Token      string
	Roles      []string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	State      SessionState
	Metadata   map[string]string
}

// IsValid returns true if the session is active and not expired.
func (s *Session) IsValid() bool {
	return s.State == SessionStateActive && time.Now().Before(s.ExpiresAt)
}

// SessionStore manages concurrent session creation, lookup, renewal, and revocation.
type SessionStore struct {
	mu            sync.RWMutex
	sessions      map[string]*Session // keyed by token
	byUserID      map[string][]string // userID -> []token
	ttl           time.Duration
	maxPerUser    int
	cleanupTicker *time.Ticker
	stopCh        chan struct{}
}

// NewSessionStore creates a session store with the given TTL and per-user session limit.
func NewSessionStore(ttl time.Duration, maxPerUser int) *SessionStore {
	s := &SessionStore{
		sessions:   make(map[string]*Session),
		byUserID:   make(map[string][]string),
		ttl:        ttl,
		maxPerUser: maxPerUser,
		stopCh:     make(chan struct{}),
	}
	return s
}

// Create creates a new session for the given user.
func (s *SessionStore) Create(userID, tenantID string, roles []string, ipAddr, userAgent string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Enforce per-user session limit
	existing := s.byUserID[userID]
	if len(existing) >= s.maxPerUser {
		// Revoke the oldest active session
		oldest := existing[0]
		if sess, ok := s.sessions[oldest]; ok {
			sess.State = SessionStateRevoked
		}
		delete(s.sessions, oldest)
		s.byUserID[userID] = existing[1:]
	}

	// Generate secure token
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	token := hex.EncodeToString(rawToken)

	now := time.Now()
	sess := &Session{
		ID:         fmt.Sprintf("sess-%s", token[:16]),
		UserID:     userID,
		TenantID:   tenantID,
		Token:      token,
		Roles:      roles,
		IPAddress:  ipAddr,
		UserAgent:  userAgent,
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(s.ttl),
		State:      SessionStateActive,
		Metadata:   make(map[string]string),
	}

	s.sessions[token] = sess
	s.byUserID[userID] = append(s.byUserID[userID], token)

	return sess, nil
}

// Get retrieves a session by token, updating LastSeenAt if valid.
func (s *SessionStore) Get(token string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if !sess.IsValid() {
		if sess.State == SessionStateActive {
			sess.State = SessionStateExpired
		}
		return nil, fmt.Errorf("session %s is %s", sess.ID, sess.State)
	}

	sess.LastSeenAt = time.Now()
	return sess, nil
}

// Renew extends the session TTL from now.
func (s *SessionStore) Renew(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return fmt.Errorf("session not found")
	}
	if sess.State != SessionStateActive {
		return fmt.Errorf("cannot renew session in state %s", sess.State)
	}

	sess.ExpiresAt = time.Now().Add(s.ttl)
	return nil
}

// Revoke invalidates a session by token.
func (s *SessionStore) Revoke(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return fmt.Errorf("session not found")
	}

	sess.State = SessionStateRevoked
	return nil
}

// RevokeAllForUser revokes all sessions belonging to the given user.
func (s *SessionStore) RevokeAllForUser(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens := s.byUserID[userID]
	count := 0
	for _, token := range tokens {
		if sess, ok := s.sessions[token]; ok && sess.State == SessionStateActive {
			sess.State = SessionStateRevoked
			count++
		}
	}
	return count
}

// PruneExpired removes all expired/revoked sessions from memory.
func (s *SessionStore) PruneExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	pruned := 0
	now := time.Now()

	for token, sess := range s.sessions {
		if sess.State != SessionStateActive || now.After(sess.ExpiresAt) {
			delete(s.sessions, token)
			// Remove from byUserID
			userTokens := s.byUserID[sess.UserID]
			for i, t := range userTokens {
				if t == token {
					s.byUserID[sess.UserID] = append(userTokens[:i], userTokens[i+1:]...)
					break
				}
			}
			pruned++
		}
	}
	return pruned
}

// ActiveCount returns the number of currently active sessions.
func (s *SessionStore) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, sess := range s.sessions {
		if sess.IsValid() {
			count++
		}
	}
	return count
}

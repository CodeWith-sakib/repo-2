package auth

import (
	"testing"
	"time"
)

func TestSessionStore_CreateAndGet(t *testing.T) {
	store := NewSessionStore(time.Hour, 3)

	sess, err := store.Create("user-1", "tenant-x", []string{"admin"}, "192.168.1.1", "curl/7.0")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if !sess.IsValid() {
		t.Error("newly created session should be valid")
	}

	fetched, err := store.Get(sess.Token)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if fetched.UserID != "user-1" {
		t.Errorf("unexpected userID: %s", fetched.UserID)
	}
}

func TestSessionStore_Revoke(t *testing.T) {
	store := NewSessionStore(time.Hour, 3)

	sess, _ := store.Create("user-2", "tenant-y", nil, "", "")
	if err := store.Revoke(sess.Token); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if _, err := store.Get(sess.Token); err == nil {
		t.Error("expected error getting revoked session")
	}
}

func TestSessionStore_RevokeAllForUser(t *testing.T) {
	store := NewSessionStore(time.Hour, 5)

	for i := 0; i < 3; i++ {
		_, _ = store.Create("user-3", "tenant-z", nil, "", "")
	}

	revoked := store.RevokeAllForUser("user-3")
	if revoked != 3 {
		t.Errorf("expected 3 revoked, got %d", revoked)
	}
	if store.ActiveCount() != 0 {
		t.Errorf("expected 0 active after revoke-all, got %d", store.ActiveCount())
	}
}

func TestSessionStore_PerUserLimit(t *testing.T) {
	store := NewSessionStore(time.Hour, 2)

	s1, _ := store.Create("user-4", "t", nil, "", "")
	s2, _ := store.Create("user-4", "t", nil, "", "")
	s3, _ := store.Create("user-4", "t", nil, "", "") // should evict s1

	// s1 should be evicted
	if _, err := store.Get(s1.Token); err == nil {
		t.Error("s1 should have been evicted by per-user limit")
	}
	// s2 and s3 should be valid
	if _, err := store.Get(s2.Token); err != nil {
		t.Errorf("s2 should still be valid: %v", err)
	}
	if _, err := store.Get(s3.Token); err != nil {
		t.Errorf("s3 should be valid: %v", err)
	}
}

func TestSessionStore_PruneExpired(t *testing.T) {
	store := NewSessionStore(20*time.Millisecond, 10)

	_, _ = store.Create("user-5", "t", nil, "", "")
	_, _ = store.Create("user-5", "t", nil, "", "")

	time.Sleep(30 * time.Millisecond)

	pruned := store.PruneExpired()
	if pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", pruned)
	}
}

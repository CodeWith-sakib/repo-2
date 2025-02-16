package auth

import (
	"testing"
	"time"
)

func TestAPIKeyLifecycle(t *testing.T) {
	store := NewAPIKeyStore()

	key, err := store.CreateKey("key-1", "org-test", RoleOperator, []string{"workflows:read", "workflows:write"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("create key failed: %v", err)
	}

	rec, err := store.Authenticate(key)
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if rec.TenantID != "org-test" || rec.Role != RoleOperator {
		t.Errorf("unexpected record: %+v", rec)
	}

	// Revoke
	if !store.Revoke("key-1") {
		t.Error("expected revoke to succeed")
	}

	if _, err := store.Authenticate(key); err == nil {
		t.Error("expected authentication failure after revocation")
	}
}

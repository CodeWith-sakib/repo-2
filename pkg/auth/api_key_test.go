package auth

import (
	"testing"
	"time"
)

func TestAPIKeyManager_GenerateAndValidate(t *testing.T) {
	mgr := NewAPIKeyManager()

	scopes := []string{"workflows:read", "workflows:write"}
	gen, err := mgr.Generate("tenant-1", "production-key", "live", scopes, time.Hour)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if gen.PlaintextKey == "" {
		t.Fatal("expected non-empty plaintext key")
	}

	meta, err := mgr.Validate(gen.PlaintextKey)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	if meta.TenantID != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", meta.TenantID)
	}

	if !meta.HasScope("workflows:read") {
		t.Error("expected hasScope workflows:read")
	}
	if meta.HasScope("admin:all") {
		t.Error("unexpected scope admin:all")
	}
}

func TestAPIKeyManager_Revocation(t *testing.T) {
	mgr := NewAPIKeyManager()

	gen, _ := mgr.Generate("tenant-2", "test-key", "test", []string{"*"}, time.Hour)
	if _, err := mgr.Validate(gen.PlaintextKey); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	if !mgr.Revoke(gen.Metadata.KeyPrefix) {
		t.Fatal("revoke should return true")
	}

	if _, err := mgr.Validate(gen.PlaintextKey); err == nil {
		t.Error("expected error after revocation")
	}
}

func TestAPIKeyManager_WildcardScope(t *testing.T) {
	meta := &APIKeyMetadata{
		Scopes: []string{"workflows:*"},
	}

	if !meta.HasScope("workflows:read") {
		t.Error("expected workflows:read to match workflows:*")
	}
	if !meta.HasScope("workflows:delete") {
		t.Error("expected workflows:delete to match workflows:*")
	}
	if meta.HasScope("storage:read") {
		t.Error("storage:read should not match workflows:*")
	}
}

package auth

import (
	"testing"
	"time"
)

func TestKeyRotationManager_Lifecycle(t *testing.T) {
	mgr, err := NewKeyRotationManager(time.Hour, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to init key manager: %v", err)
	}

	k1, err := mgr.GetCurrentKey()
	if err != nil {
		t.Fatalf("failed to get current key: %v", err)
	}

	if !mgr.VerifyKey(k1.ID, k1.Secret) {
		t.Errorf("expected k1 secret to verify successfully")
	}

	// Rotate
	if err := mgr.Rotate(); err != nil {
		t.Fatalf("rotation failed: %v", err)
	}

	k2, _ := mgr.GetCurrentKey()
	if k2.ID == k1.ID {
		t.Errorf("expected new key ID after rotation")
	}

	// k1 still valid during grace overlap
	if !mgr.VerifyKey(k1.ID, k1.Secret) {
		t.Errorf("k1 should still verify during grace overlap")
	}

	// Wait for grace overlap to expire
	time.Sleep(120 * time.Millisecond)

	if mgr.VerifyKey(k1.ID, k1.Secret) {
		t.Errorf("k1 should have expired after grace window")
	}

	pruned := mgr.PruneExpired()
	if pruned != 1 {
		t.Errorf("expected 1 pruned key, got %d", pruned)
	}
}

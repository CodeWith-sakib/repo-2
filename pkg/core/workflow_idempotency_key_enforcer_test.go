package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIdempotencyKeyEnforcer(t *testing.T) {
	enforcer := NewIdempotencyKeyEnforcer(1 * time.Hour)

	wfID, err := enforcer.RegisterOrCheck(context.Background(), "req-abc-123", "wf-999")
	if err != nil {
		t.Fatalf("first registration should succeed, got %v", err)
	}
	if wfID != "wf-999" {
		t.Errorf("expected wf-999, got %s", wfID)
	}

	// Duplicate check
	existingID, err := enforcer.RegisterOrCheck(context.Background(), "req-abc-123", "wf-1000")
	if !errors.Is(err, ErrDuplicateIdempotencyKey) {
		t.Fatalf("expected ErrDuplicateIdempotencyKey, got %v", err)
	}
	if existingID != "wf-999" {
		t.Errorf("expected to return existing wf-999, got %s", existingID)
	}

	// Expired purge
	future := time.Now().Add(2 * time.Hour)
	purged := enforcer.PurgeExpired(future)
	if purged != 1 {
		t.Errorf("expected 1 record purged, got %d", purged)
	}
}

package auth

import (
	"testing"
)

func TestAuditLoggerSigningAndVerification(t *testing.T) {
	logger := NewAuditLogger("secret-audit-key-1234")

	evt, err := logger.Log("user-alex", AuditRunTriggered, "run-42", "Triggered via CLI")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if evt.Signature == "" {
		t.Fatal("expected non-empty cryptographic signature")
	}

	if !logger.Verify(evt) {
		t.Error("expected valid signature verification")
	}

	// Tampered event must fail verification
	evt.Details = "Tampered details"
	if logger.Verify(evt) {
		t.Error("expected verification to fail for tampered event")
	}
}

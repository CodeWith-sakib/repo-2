package auth

import (
	"testing"
	"time"
)

func TestWebhookSignature_SignAndVerify(t *testing.T) {
	verifier := NewWebhookSignatureVerifier(5 * time.Minute)

	payload := []byte(`{"event":"workflow.completed","run_id":"run-123"}`)
	secret := "whsec_super_secret_signing_key_42"
	now := time.Now()

	header := SignHeader(payload, secret, now)

	// Valid verification
	if err := verifier.Verify(payload, header, secret, now); err != nil {
		t.Fatalf("verification failed on valid signature: %v", err)
	}

	// Payload tampering -> should fail
	tampered := []byte(`{"event":"workflow.completed","run_id":"run-tampered"}`)
	if err := verifier.Verify(tampered, header, secret, now); err == nil {
		t.Error("expected error on tampered payload")
	}

	// Wrong secret -> should fail
	if err := verifier.Verify(payload, header, "wrong_secret", now); err == nil {
		t.Error("expected error with wrong secret")
	}

	// Replay beyond tolerance -> should fail
	futureTime := now.Add(10 * time.Minute)
	if err := verifier.Verify(payload, header, secret, futureTime); err == nil {
		t.Error("expected replay failure outside tolerance window")
	}
}

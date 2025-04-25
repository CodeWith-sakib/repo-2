package webhook

import (
	"testing"
)

func TestPayloadSigner(t *testing.T) {
	signer256 := NewPayloadSigner("wh-secret-key", AlgorithmSHA256)
	payload := []byte(`{"event":"run.completed","id":"123"}`)

	sig256, err := signer256.Sign(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !signer256.Verify(payload, sig256) {
		t.Error("expected valid verification for sha256 signature")
	}

	signer512 := NewPayloadSigner("wh-secret-key", AlgorithmSHA512)
	sig512, err := signer512.Sign(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !signer512.Verify(payload, sig512) {
		t.Error("expected valid verification for sha512 signature")
	}

	if signer256.Verify(payload, sig512) {
		t.Error("sha256 signer should not verify sha512 signature")
	}
}

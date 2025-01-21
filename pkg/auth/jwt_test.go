package auth

import (
	"testing"
	"time"
)

func TestJWTManagerSignAndVerify(t *testing.T) {
	mgr := NewJWTManager("super-secret-key-12345")

	claims := Claims{
		Subject:   "user-alex",
		Role:      RoleAdmin,
		TenantID:  "tenant-core",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := mgr.Sign(claims)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	verified, err := mgr.Verify(token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	if verified.Subject != "user-alex" || verified.Role != RoleAdmin {
		t.Errorf("claims mismatch: %+v", verified)
	}

	// Verify tampering detection
	tampered := token + "bad"
	if _, err := mgr.Verify(tampered); err == nil {
		t.Error("expected verification error on tampered token")
	}
}

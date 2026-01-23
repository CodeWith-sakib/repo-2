package auth

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
)

func TestMutualTLSVerifier(t *testing.T) {
	policy := MutualTLSPolicy{
		AllowedCommonNames: []string{"worker-agent.kestrelflow.io"},
		AllowedSANs:        []string{"agent.internal.net"},
		RequireClientCert:  true,
	}

	verifier := NewMutualTLSVerifier(policy)

	validCert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "worker-agent.kestrelflow.io",
		},
		DNSNames: []string{"agent.internal.net"},
	}

	if err := verifier.VerifyPeerCertificate(validCert); err != nil {
		t.Fatalf("unexpected error validating valid cert: %v", err)
	}

	// CN mismatch
	invalidCN := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "unauthorized.io",
		},
		DNSNames: []string{"agent.internal.net"},
	}
	if err := verifier.VerifyPeerCertificate(invalidCN); err == nil {
		t.Error("expected error for unauthorized CN, got nil")
	}

	// Missing cert when required
	if err := verifier.VerifyPeerCertificate(nil); err == nil {
		t.Error("expected error for nil certificate when required")
	}
}

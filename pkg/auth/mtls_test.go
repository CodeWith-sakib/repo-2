package auth

import (
	"testing"
	"time"
)

func TestGenerateSelfSignedCert_ValidateAndLoad(t *testing.T) {
	cfg := MTLSCertConfig{
		CommonName:   "kestrelflow-test-ca",
		Organization: []string{"KestrelFlow Inc."},
		DNSNames:     []string{"localhost", "kestrelflow.internal"},
		ValidFor:     24 * time.Hour,
		IsCA:         true,
	}

	gen, err := GenerateSelfSignedCert(cfg)
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert failed: %v", err)
	}

	if len(gen.CertPEM) == 0 {
		t.Error("cert PEM must not be empty")
	}
	if len(gen.KeyPEM) == 0 {
		t.Error("key PEM must not be empty")
	}

	tlsCert, err := LoadTLSCertificate(gen.CertPEM, gen.KeyPEM)
	if err != nil {
		t.Fatalf("LoadTLSCertificate failed: %v", err)
	}
	if tlsCert.Leaf == nil && len(tlsCert.Certificate) == 0 {
		t.Error("expected certificate chain to be non-empty")
	}

	pool, err := ParseCertPool(gen.CertPEM)
	if err != nil {
		t.Fatalf("ParseCertPool failed: %v", err)
	}

	cert, err := ValidateCertificate(gen.CertPEM, pool)
	if err != nil {
		t.Fatalf("ValidateCertificate failed: %v", err)
	}
	if cert.Subject.CommonName != cfg.CommonName {
		t.Errorf("unexpected CN: %s", cert.Subject.CommonName)
	}
}

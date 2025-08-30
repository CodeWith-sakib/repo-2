package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// MTLSCertConfig defines generation parameters for a self-signed mTLS certificate.
type MTLSCertConfig struct {
	CommonName   string
	Organization []string
	DNSNames     []string
	IPAddresses  []net.IP
	ValidFor     time.Duration
	IsCA         bool
}

// GeneratedCert holds the PEM-encoded certificate and private key.
type GeneratedCert struct {
	CertPEM []byte
	KeyPEM  []byte
}

// GenerateSelfSignedCert creates an ECDSA P-256 self-signed certificate.
func GenerateSelfSignedCert(cfg MTLSCertConfig) (*GeneratedCert, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generate serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:    cfg.DNSNames,
		IPAddresses: cfg.IPAddresses,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(cfg.ValidFor),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	if cfg.IsCA {
		template.IsCA = true
		template.BasicConstraintsValid = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("create cert: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return &GeneratedCert{CertPEM: certPEM, KeyPEM: keyPEM}, nil
}

// LoadTLSCertificate loads a tls.Certificate from PEM-encoded cert and key bytes.
func LoadTLSCertificate(certPEM, keyPEM []byte) (tls.Certificate, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load tls certificate: %w", err)
	}
	return cert, nil
}

// BuildMTLSServerConfig creates a tls.Config for mutual TLS (mTLS) server use.
// It requires client certificates signed by the given CA pool.
func BuildMTLSServerConfig(serverCert tls.Certificate, caPool *x509.CertPool) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}
}

// BuildMTLSClientConfig creates a tls.Config for mTLS client use.
func BuildMTLSClientConfig(clientCert tls.Certificate, caPool *x509.CertPool, serverName string) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}
}

// ParseCertPool creates an x509.CertPool from PEM-encoded CA certificate bytes.
func ParseCertPool(caPEM []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	return pool, nil
}

// ValidateCertificate verifies a PEM cert against a CA pool and returns subject metadata.
func ValidateCertificate(certPEM []byte, caPool *x509.CertPool) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}

	opts := x509.VerifyOptions{
		Roots: caPool,
	}
	if _, err := cert.Verify(opts); err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	return cert, nil
}

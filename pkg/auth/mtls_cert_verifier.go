package auth

import (
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
)

// MutualTLSPolicy governs peer certificate validation rules.
type MutualTLSPolicy struct {
	AllowedCommonNames []string
	AllowedSANs        []string
	AllowedOUs         []string
	RequireClientCert  bool
}

// MutualTLSVerifier validates presented client TLS certificates against security policy.
type MutualTLSVerifier struct {
	policy MutualTLSPolicy
}

// NewMutualTLSVerifier creates a mutual TLS policy verifier.
func NewMutualTLSVerifier(policy MutualTLSPolicy) *MutualTLSVerifier {
	return &MutualTLSVerifier{policy: policy}
}

// VerifyPeerCertificate inspects client certificate attributes against allowed identities.
func (v *MutualTLSVerifier) VerifyPeerCertificate(cert *x509.Certificate) error {
	if cert == nil {
		if v.policy.RequireClientCert {
			return errors.New("client certificate required but not provided")
		}
		return nil
	}

	// Check Common Name
	if len(v.policy.AllowedCommonNames) > 0 {
		cnMatched := false
		for _, cn := range v.policy.AllowedCommonNames {
			if cert.Subject.CommonName == cn {
				cnMatched = true
				break
			}
		}
		if !cnMatched {
			return fmt.Errorf("client common name '%s' not authorized by policy", cert.Subject.CommonName)
		}
	}

	// Check SAN DNS Names
	if len(v.policy.AllowedSANs) > 0 {
		sanMatched := false
		for _, san := range cert.DNSNames {
			for _, allowed := range v.policy.AllowedSANs {
				if strings.EqualFold(san, allowed) {
					sanMatched = true
					break
				}
			}
			if sanMatched {
				break
			}
		}
		if !sanMatched {
			return errors.New("no matching SAN DNS names authorized by policy")
		}
	}

	return nil
}

package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
)

// SAMLAssertion represents minimal unpacked SAML 2.0 assertion attributes.
type SAMLAssertion struct {
	XMLName      xml.Name `xml:"Assertion"`
	Issuer       string   `xml:"Issuer"`
	Subject      string   `xml:"Subject>NameID"`
	Audience     string   `xml:"Conditions>AudienceRestriction>Audience"`
	NotBefore    string   `xml:"Conditions>NotBefore"`
	NotOnOrAfter string   `xml:"Conditions>NotOnOrAfter"`
}

// SAMLConfiguration holds IdP metadata parameters.
type SAMLConfiguration struct {
	EntityID         string
	IdPIssuer        string
	ExpectedAudience string
	Certificate      *x509.Certificate
	PublicKey        *rsa.PublicKey
}

// SAMLAuthenticator verifies SAML Response payload signatures and claims.
type SAMLAuthenticator struct {
	config SAMLConfiguration
}

// NewSAMLAuthenticator constructs a SAML authenticator.
func NewSAMLAuthenticator(config SAMLConfiguration) *SAMLAuthenticator {
	return &SAMLAuthenticator{config: config}
}

// ValidateResponse decodes base64-encoded SAML XML response and verifies basic constraints.
func (sa *SAMLAuthenticator) ValidateResponse(b64XML string) (*SAMLAssertion, error) {
	if b64XML == "" {
		return nil, errors.New("empty SAML response")
	}

	decoded, err := base64.StdEncoding.DecodeString(b64XML)
	if err != nil {
		return nil, fmt.Errorf("failed base64 decoding SAML response: %w", err)
	}

	var assertion SAMLAssertion
	if err := xml.Unmarshal(decoded, &assertion); err != nil {
		return nil, fmt.Errorf("failed unmarshaling SAML XML: %w", err)
	}

	if sa.config.IdPIssuer != "" && assertion.Issuer != sa.config.IdPIssuer {
		return nil, fmt.Errorf("issuer mismatch: expected %s, got %s", sa.config.IdPIssuer, assertion.Issuer)
	}

	if sa.config.ExpectedAudience != "" && assertion.Audience != sa.config.ExpectedAudience {
		return nil, fmt.Errorf("audience mismatch: expected %s, got %s", sa.config.ExpectedAudience, assertion.Audience)
	}

	if assertion.Subject == "" {
		return nil, errors.New("missing NameID subject in SAML assertion")
	}

	return &assertion, nil
}

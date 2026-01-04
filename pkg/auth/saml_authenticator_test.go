package auth

import (
	"encoding/base64"
	"testing"
)

func TestSAMLAuthenticator(t *testing.T) {
	cfg := SAMLConfiguration{
		EntityID:         "https://kestrelflow.io/saml/sp",
		IdPIssuer:        "https://idp.okta.com/exk101",
		ExpectedAudience: "https://kestrelflow.io/saml/sp",
	}

	auth := NewSAMLAuthenticator(cfg)

	xmlData := `<Assertion>
		<Issuer>https://idp.okta.com/exk101</Issuer>
		<Subject>
			<NameID>operator@kestrelflow.io</NameID>
		</Subject>
		<Conditions>
			<AudienceRestriction>
				<Audience>https://kestrelflow.io/saml/sp</Audience>
			</AudienceRestriction>
		</Conditions>
	</Assertion>`

	b64 := base64.StdEncoding.EncodeToString([]byte(xmlData))
	assertion, err := auth.ValidateResponse(b64)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if assertion.Subject != "operator@kestrelflow.io" {
		t.Errorf("unexpected subject: %s", assertion.Subject)
	}

	// Test issuer mismatch
	badIssuerXML := `<Assertion>
		<Issuer>https://evil-idp.com</Issuer>
		<Subject><NameID>hacker</NameID></Subject>
		<Conditions><AudienceRestriction><Audience>https://kestrelflow.io/saml/sp</Audience></AudienceRestriction></Conditions>
	</Assertion>`
	badB64 := base64.StdEncoding.EncodeToString([]byte(badIssuerXML))
	_, err = auth.ValidateResponse(badB64)
	if err == nil {
		t.Error("expected issuer validation failure, got nil")
	}
}

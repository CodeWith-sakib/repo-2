package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// WebhookSignatureVerifier verifies HMAC-SHA256 signatures with replay protection.
type WebhookSignatureVerifier struct {
	tolerance time.Duration
}

// NewWebhookSignatureVerifier creates a verifier with tolerance window.
func NewWebhookSignatureVerifier(tolerance time.Duration) *WebhookSignatureVerifier {
	if tolerance <= 0 {
		tolerance = 5 * time.Minute
	}
	return &WebhookSignatureVerifier{
		tolerance: tolerance,
	}
}

// SignHeader computes an RFC-compliant signature header: t=<unix_sec>,v1=<hex_sig>.
func SignHeader(payload []byte, secret string, now time.Time) string {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	toSign := fmt.Sprintf("%s.%s", timestamp, string(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(toSign))
	sigHex := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("t=%s,v1=%s", timestamp, sigHex)
}

// Verify checks the signature header against the payload and secret.
func (v *WebhookSignatureVerifier) Verify(payload []byte, signatureHeader string, secret string, now time.Time) error {
	if signatureHeader == "" {
		return fmt.Errorf("missing signature header")
	}

	parts := strings.Split(signatureHeader, ",")
	var timestampStr, receivedSig string

	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestampStr = kv[1]
		case "v1":
			receivedSig = kv[1]
		}
	}

	if timestampStr == "" || receivedSig == "" {
		return fmt.Errorf("invalid signature header format")
	}

	tsUnix, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp in header: %w", err)
	}

	headerTime := time.Unix(tsUnix, 0)
	diff := now.Sub(headerTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > v.tolerance {
		return fmt.Errorf("webhook timestamp outside tolerance window (%v > %v)", diff, v.tolerance)
	}

	toSign := fmt.Sprintf("%s.%s", timestampStr, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(toSign))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(receivedSig)) != 1 {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

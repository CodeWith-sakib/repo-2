package core

import (
	"strings"
	"testing"
)

func TestSensitiveDataScrubber(t *testing.T) {
	scrubber := NewSensitiveDataScrubber()

	text := "Customer card 4111-2222-3333-4444 and ssn 000-12-3456 with Bearer abcdef1234567890abcdef"
	cleaned := scrubber.ScrubText(text)

	if strings.Contains(cleaned, "4111-2222") {
		t.Error("credit card was not scrubbed")
	}
	if strings.Contains(cleaned, "000-12-3456") {
		t.Error("ssn was not scrubbed")
	}
	if strings.Contains(cleaned, "abcdef1234567890") {
		t.Error("bearer token was not scrubbed")
	}

	// Map scrubbing
	m := map[string]interface{}{
		"api_key_secret": "raw-super-secret",
		"normal_info":    "safe value",
	}
	cleanedMap := scrubber.ScrubMap(m)
	if cleanedMap["api_key_secret"] != "[REDACTED]" {
		t.Errorf("expected secret redacted, got %v", cleanedMap["api_key_secret"])
	}
	if cleanedMap["normal_info"] != "safe value" {
		t.Errorf("expected normal_info preserved, got %v", cleanedMap["normal_info"])
	}
}

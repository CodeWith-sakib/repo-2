package core

import (
	"testing"
)

func TestFormatRegistryValidation(t *testing.T) {
	r := NewFormatRegistry()

	cases := []struct {
		format string
		val    string
		valid  bool
	}{
		{"uuid", "123e4567-e89b-12d3-a456-426614174000", true},
		{"uuid", "invalid-uuid", false},
		{"email", "user@example.com", true},
		{"email", "invalid-email@", false},
		{"uri", "https://kestrelflow.io/docs", true},
		{"uri", "not a uri", false},
		{"ipv4", "192.168.1.1", true},
		{"ipv4", "999.999.999.999", false},
		{"ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"ipv6", "192.168.1.1", false},
		{"date-time", "2026-05-01T12:00:00Z", true},
		{"date-time", "2026-05-01", false},
		{"duration", "15m", true},
		{"duration", "invalid", false},
		{"base64", "SGVsbG8gV29ybGQ=", true},
		{"base64", "%%%notbase64", false},
	}

	for _, tc := range cases {
		err := r.Validate(tc.format, tc.val)
		if tc.valid && err != nil {
			t.Errorf("format %s val %q expected valid, got error: %v", tc.format, tc.val, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("format %s val %q expected invalid, got nil error", tc.format, tc.val)
		}
	}
}

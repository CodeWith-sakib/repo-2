package types

import (
	"testing"
)

func TestStringSanitizer(t *testing.T) {
	s := NewStringSanitizer()
	res := s.Sanitize("  <script>alert(1)</script>  ")
	expected := "&lt;script&gt;alert(1)&lt;/script&gt;"
	if res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}

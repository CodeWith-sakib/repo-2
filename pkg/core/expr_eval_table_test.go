package core

import (
	"testing"
)

func TestCoercions(t *testing.T) {
	if !CoerceToBool("yes") {
		t.Error("expected yes to coerce to true")
	}
	if CoerceToBool("0") {
		t.Error("expected 0 to coerce to false")
	}

	f, ok := CoerceToFloat("123.45")
	if !ok || f != 123.45 {
		t.Errorf("expected 123.45, got %v", f)
	}

	s := CoerceToString(99)
	if s != "99" {
		t.Errorf("expected 99, got %s", s)
	}
}

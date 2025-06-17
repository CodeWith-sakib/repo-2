package dashboard

import (
	"testing"
)

func TestViewModelFilter(t *testing.T) {
	f := NewViewModelFilter("COMPLETED")
	if !f.Match("COMPLETED") {
		t.Error("expected match")
	}
	if f.Match("FAILED") {
		t.Error("expected non-match")
	}
}

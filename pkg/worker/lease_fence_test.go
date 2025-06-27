package worker

import "testing"

func TestFencedLeaseToken(t *testing.T) {
	tok := NewFencedLeaseToken(100)
	v := tok.NextToken()
	if v != 101 {
		t.Errorf("expected 101, got %d", v)
	}
	if !tok.ValidateToken(101) {
		t.Error("expected token 101 to validate")
	}
}

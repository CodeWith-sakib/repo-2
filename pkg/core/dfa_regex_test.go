package core

import (
	"testing"
)

func TestDFAMachine_Identifier(t *testing.T) {
	dfa, err := CompileIdentifierPattern()
	if err != nil {
		t.Fatalf("failed to compile identifier dfa: %v", err)
	}

	valid := []string{"foo", "Bar", "_custom_123", "a", "_"}
	for _, v := range valid {
		if !dfa.Match(v) {
			t.Errorf("expected %q to match identifier pattern", v)
		}
	}

	invalid := []string{"123foo", "", "foo bar", "test-dash", "a.b"}
	for _, inv := range invalid {
		if dfa.Match(inv) {
			t.Errorf("expected %q to NOT match identifier pattern", inv)
		}
	}
}

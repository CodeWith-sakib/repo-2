package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLIBasics(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	c := NewCLI(out, errOut)

	// Version
	code := c.Execute([]string{"version"})
	if code != 0 || !strings.Contains(out.String(), "kestrelflow v1.0.0") {
		t.Fatalf("expected version output, got code %d: %s", code, out.String())
	}

	// Usage / Help
	out.Reset()
	code = c.Execute([]string{"help"})
	if code != 0 || !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected usage output, got code %d: %s", code, out.String())
	}

	// Workflow commands
	out.Reset()
	code = c.Execute([]string{"workflow", "list"})
	if code != 0 || !strings.Contains(out.String(), "Listing workflows") {
		t.Fatalf("expected workflow list output, got code %d: %s", code, out.String())
	}

	// Run commands
	out.Reset()
	code = c.Execute([]string{"run", "get", "run-123"})
	if code != 0 || !strings.Contains(out.String(), "Run status for run-123") {
		t.Fatalf("expected run status output, got code %d: %s", code, out.String())
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestSimpleTableWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewSimpleTableWriter(&buf)
	w.RenderHeader("ID", "STATE")
	w.RenderRow("run-1", "COMPLETED")
	output := buf.String()
	if !strings.Contains(output, "ID	STATE") || !strings.Contains(output, "run-1	COMPLETED") {
		t.Errorf("unexpected table output: %s", output)
	}
}

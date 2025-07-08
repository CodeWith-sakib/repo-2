package shell

import (
	"testing"
)

func TestEnvFilterWhitelist(t *testing.T) {
	f := NewEnvFilterWhitelist([]string{"PATH", "USER"})
	res := f.Filter([]string{"PATH=/bin", "SECRET=123", "USER=alex"})
	if len(res) != 2 {
		t.Errorf("expected 2 items, got %d", len(res))
	}
}

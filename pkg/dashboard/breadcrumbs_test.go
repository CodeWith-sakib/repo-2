package dashboard

import (
	"strings"
	"testing"
)

func TestBreadcrumbTrail(t *testing.T) {
	b := NewBreadcrumbTrail()
	b.Add("Workflows", "/workflows")
	html := b.RenderHTML()
	if !strings.Contains(html, "<a href='/'>Home</a>") || !strings.Contains(html, "<a href='/workflows'>Workflows</a>") {
		t.Errorf("unexpected html: %s", html)
	}
}

package types

import (
	"html"
	"strings"
)

type StringSanitizer struct{}

func NewStringSanitizer() *StringSanitizer {
	return &StringSanitizer{}
}

func (s *StringSanitizer) Sanitize(input string) string {
	trimmed := strings.TrimSpace(input)
	return html.EscapeString(trimmed)
}

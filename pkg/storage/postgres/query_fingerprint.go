package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	singleQuoteLiteralRegex = regexp.MustCompile(`'[^']*'`)
	numericLiteralRegex     = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	multipleSpacesRegex     = regexp.MustCompile(`\s+`)
)

// QueryFingerprinter normalizes dynamic SQL queries into canonical fingerprints to group execution metrics.
type QueryFingerprinter struct{}

// NewQueryFingerprinter creates a query fingerprinter.
func NewQueryFingerprinter() *QueryFingerprinter {
	return &QueryFingerprinter{}
}

// Normalize strips literal constants, quoted strings, and formats whitespace uniformly.
func (f *QueryFingerprinter) Normalize(sqlQuery string) string {
	s := strings.TrimSpace(sqlQuery)
	// Replace string literals with ?
	s = singleQuoteLiteralRegex.ReplaceAllString(s, "?")
	// Replace numeric literals with ?
	s = numericLiteralRegex.ReplaceAllString(s, "?")
	// Collapse multiple whitespaces
	s = multipleSpacesRegex.ReplaceAllString(s, " ")
	return strings.ToLower(s)
}

// Fingerprint calculates a 16-character hex hash of the normalized query string.
func (f *QueryFingerprinter) Fingerprint(sqlQuery string) string {
	norm := f.Normalize(sqlQuery)
	h := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(h[:8]) // 16 hex chars
}

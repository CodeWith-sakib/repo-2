package core

import (
	"regexp"
	"strings"
	"sync"
)

var (
	creditCardRegex = regexp.MustCompile(`\b(?:\d{4}[ -]?){3}\d{4}\b`)
	ssnRegex        = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	bearerAuthRegex = regexp.MustCompile(`(?i)(bearer\s+[a-zA-Z0-9_\-\.]{16,})`)
)

// SensitiveDataScrubber masks confidential PII, secrets, and auth tokens from workflow audit logs.
type SensitiveDataScrubber struct {
	mu            sync.RWMutex
	customRegexes []*regexp.Regexp
	maskPattern   string
}

// NewSensitiveDataScrubber creates a data scrubbing utility.
func NewSensitiveDataScrubber() *SensitiveDataScrubber {
	return &SensitiveDataScrubber{
		maskPattern: "[REDACTED]",
	}
}

// AddCustomPattern registers an additional sensitive regex pattern to scrub.
func (s *SensitiveDataScrubber) AddCustomPattern(pattern string) error {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.customRegexes = append(s.customRegexes, compiled)
	return nil
}

// ScrubText replaces all discovered credit card, SSN, bearer token, and custom patterns with [REDACTED].
func (s *SensitiveDataScrubber) ScrubText(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := creditCardRegex.ReplaceAllString(input, s.maskPattern)
	out = ssnRegex.ReplaceAllString(out, s.maskPattern)
	out = bearerAuthRegex.ReplaceAllString(out, s.maskPattern)

	for _, re := range s.customRegexes {
		out = re.ReplaceAllString(out, s.maskPattern)
	}

	return out
}

// ScrubMap recursively scrubs string values within a nested map structure.
func (s *SensitiveDataScrubber) ScrubMap(data map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		lowerKey := strings.ToLower(k)
		if strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "token") {
			out[k] = s.maskPattern
			continue
		}

		switch val := v.(type) {
		case string:
			out[k] = s.ScrubText(val)
		case map[string]interface{}:
			out[k] = s.ScrubMap(val)
		default:
			out[k] = val
		}
	}
	return out
}

package core

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)
)

type FormatValidator func(value string) bool

type FormatRegistry struct {
	validators map[string]FormatValidator
}

func NewFormatRegistry() *FormatRegistry {
	r := &FormatRegistry{
		validators: make(map[string]FormatValidator),
	}
	r.registerBuiltins()
	return r
}

func (r *FormatRegistry) Register(format string, fn FormatValidator) {
	r.validators[strings.ToLower(format)] = fn
}

func (r *FormatRegistry) Validate(format, value string) error {
	fn, exists := r.validators[strings.ToLower(format)]
	if !exists {
		return nil
	}
	if !fn(value) {
		return fmt.Errorf("%w: value %q does not match format %s", ErrValidationFailed, value, format)
	}
	return nil
}

func (r *FormatRegistry) registerBuiltins() {
	r.Register("uuid", func(val string) bool {
		return uuidRegex.MatchString(val)
	})

	r.Register("email", func(val string) bool {
		_, err := mail.ParseAddress(val)
		return err == nil
	})

	r.Register("uri", func(val string) bool {
		u, err := url.ParseRequestURI(val)
		return err == nil && u.Scheme != "" && u.Host != ""
	})

	r.Register("ipv4", func(val string) bool {
		ip := net.ParseIP(val)
		return ip != nil && strings.Contains(val, ".")
	})

	r.Register("ipv6", func(val string) bool {
		ip := net.ParseIP(val)
		return ip != nil && strings.Contains(val, ":")
	})

	r.Register("date-time", func(val string) bool {
		_, err := time.Parse(time.RFC3339, val)
		return err == nil
	})

	r.Register("duration", func(val string) bool {
		_, err := time.ParseDuration(val)
		return err == nil
	})

	r.Register("base64", func(val string) bool {
		_, err := base64.StdEncoding.DecodeString(val)
		return err == nil
	})

	r.Register("hostname", func(val string) bool {
		return len(val) <= 253 && hostnameRegex.MatchString(val)
	})
}

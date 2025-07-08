package shell

import (
	"strings"
)

type EnvFilterWhitelist struct {
	allowed []string
}

func NewEnvFilterWhitelist(allowed []string) *EnvFilterWhitelist {
	return &EnvFilterWhitelist{allowed: allowed}
}

func (ef *EnvFilterWhitelist) Filter(env []string) []string {
	var filtered []string
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) > 0 {
			for _, a := range ef.allowed {
				if parts[0] == a {
					filtered = append(filtered, e)
					break
				}
			}
		}
	}
	return filtered
}

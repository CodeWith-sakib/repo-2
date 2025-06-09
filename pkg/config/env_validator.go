package config

import (
	"fmt"
	"os"
)

type EnvValidator struct{}

func NewEnvValidator() *EnvValidator {
	return &EnvValidator{}
}

func (v *EnvValidator) Require(keys ...string) error {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return fmt.Errorf("required environment variable %s is not set", k)
		}
	}
	return nil
}

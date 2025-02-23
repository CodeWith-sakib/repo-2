package core

import (
	"testing"
	"time"
)

func TestCronSpecMatches(t *testing.T) {
	sched, err := ParseCronSpec("* * * * *")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	now := time.Now().UTC()
	if !sched.Matches(now) {
		t.Error("wildcard should match current time")
	}
}

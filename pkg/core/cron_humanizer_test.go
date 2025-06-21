package core

import "testing"

func TestCronHumanizer(t *testing.T) {
	h := NewCronHumanizer()
	if h.Humanize("0 0 * * *") != "Every day at midnight" {
		t.Error("unexpected humanized string")
	}
}

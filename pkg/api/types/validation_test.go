package types

import (
	"testing"
)

type sampleRequest struct {
	Name  string `validate:"required,identifier,min=3,max=30"`
	Email string `validate:"email"`
	Count int    `validate:"min=1,max=100"`
	URL   string `validate:"url"`
}

func TestStructValidator(t *testing.T) {
	v := NewStructValidator()

	valid := sampleRequest{
		Name:  "my-workflow-1",
		Email: "admin@kestrelflow.io",
		Count: 5,
		URL:   "https://kestrelflow.io/api/v1",
	}

	errs := v.Validate(valid)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got: %v", errs)
	}

	invalid := sampleRequest{
		Name:  "no",
		Email: "invalid-email",
		Count: 0,
		URL:   "not-a-url",
	}

	errs = v.Validate(invalid)
	if len(errs) != 4 {
		t.Fatalf("expected 4 validation errors, got %d: %v", len(errs), errs)
	}
}

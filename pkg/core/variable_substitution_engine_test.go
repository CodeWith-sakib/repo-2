package core

import (
	"testing"
)

func TestVariableSubstitutionEngine(t *testing.T) {
	engine := NewVariableSubstitutionEngine()
	env := map[string]string{
		"env":     "production",
		"region":  "us-east-1",
		"cluster": "k8s-prod-01",
	}

	tpl := "Deploying to ${cluster}.${region} in environment ${env}"
	res, err := engine.SubstituteString(tpl, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Deploying to k8s-prod-01.us-east-1 in environment production"
	if res != expected {
		t.Errorf("expected '%s', got '%s'", expected, res)
	}

	// Missing variable
	badTpl := "Deploying to ${unknown_host}"
	_, err = engine.SubstituteString(badTpl, env)
	if err == nil {
		t.Error("expected error for missing variable, got nil")
	}

	// Map substitution
	m := map[string]string{
		"host": "api.${env}.company.com",
		"tag":  "release-${env}",
	}
	mRes, err := engine.SubstituteMap(m, env)
	if err != nil {
		t.Fatalf("unexpected map substitution error: %v", err)
	}
	if mRes["host"] != "api.production.company.com" {
		t.Errorf("unexpected host: %s", mRes["host"])
	}
}

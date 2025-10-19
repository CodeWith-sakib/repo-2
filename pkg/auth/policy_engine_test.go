package auth

import (
	"testing"
)

func TestABACPolicyEngine_AllowAndExplicitDeny(t *testing.T) {
	engine := NewABACPolicyEngine()

	// Rule 1: Allow operators to read workflows
	_ = engine.AddRule(&PolicyRule{
		ID:        "allow-operator-read",
		Effect:    EffectAllow,
		Roles:     []string{"operator"},
		Actions:   []string{"read"},
		Resources: []string{"workflow:*"},
	})

	// Rule 2: Explicitly deny access to secret workflow
	_ = engine.AddRule(&PolicyRule{
		ID:        "deny-secret-workflow",
		Effect:    EffectDeny,
		Roles:     []string{"*"},
		Actions:   []string{"*"},
		Resources: []string{"workflow:secret-vault"},
	})

	// Operator reading standard workflow -> ALLOW
	allowed, _ := engine.Evaluate(AccessRequest{
		TenantID: "corp-a",
		Roles:    []string{"operator"},
		Action:   "read",
		Resource: "workflow:billing-job",
	})
	if !allowed {
		t.Error("expected allow for operator reading normal workflow")
	}

	// Operator reading secret workflow -> DENY (explicit deny overrides allow)
	denied, reason := engine.Evaluate(AccessRequest{
		TenantID: "corp-a",
		Roles:    []string{"operator"},
		Action:   "read",
		Resource: "workflow:secret-vault",
	})
	if denied {
		t.Errorf("expected deny for secret vault, got allow (reason: %s)", reason)
	}

	// Operator attempting write -> DENY (default deny)
	writeDenied, _ := engine.Evaluate(AccessRequest{
		TenantID: "corp-a",
		Roles:    []string{"operator"},
		Action:   "write",
		Resource: "workflow:billing-job",
	})
	if writeDenied {
		t.Error("expected default deny for write")
	}
}

package auth

import (
	"fmt"
	"strings"
	"sync"
)

// PolicyEffect is the decision outcome of a policy rule.
type PolicyEffect string

const (
	EffectAllow PolicyEffect = "ALLOW"
	EffectDeny  PolicyEffect = "DENY"
)

// PolicyRule defines an individual ABAC rule.
type PolicyRule struct {
	ID        string       `json:"id"`
	Effect    PolicyEffect `json:"effect"`
	Roles     []string     `json:"roles"`     // empty = match any role
	Actions   []string     `json:"actions"`   // e.g. "read", "write", "*"
	Resources []string     `json:"resources"` // e.g. "workflow:*", "run:123"
	Tenants   []string     `json:"tenants"`   // empty = match any tenant
}

// AccessRequest encapsulates the context of an access authorization check.
type AccessRequest struct {
	TenantID string
	Roles    []string
	Action   string
	Resource string
}

// ABACPolicyEngine evaluates authorization decisions using attribute-based policy rules.
type ABACPolicyEngine struct {
	mu    sync.RWMutex
	rules []*PolicyRule
}

// NewABACPolicyEngine creates an empty ABAC engine.
func NewABACPolicyEngine() *ABACPolicyEngine {
	return &ABACPolicyEngine{}
}

// AddRule registers a policy rule.
func (e *ABACPolicyEngine) AddRule(rule *PolicyRule) error {
	if rule == nil || rule.ID == "" {
		return fmt.Errorf("invalid rule: nil or empty ID")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
	return nil
}

// Evaluate evaluates access against all rules.
// Semantics:
// 1. If any matching rule has effect DENY -> DENY (explicit deny overrides).
// 2. If any matching rule has effect ALLOW -> ALLOW.
// 3. Otherwise -> DENY (default deny).
func (e *ABACPolicyEngine) Evaluate(req AccessRequest) (bool, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	matchedAllow := false
	var allowReason string

	for _, rule := range e.rules {
		if !e.matches(rule, req) {
			continue
		}

		if rule.Effect == EffectDeny {
			return false, fmt.Sprintf("explicit deny by rule %q", rule.ID)
		}

		if rule.Effect == EffectAllow {
			matchedAllow = true
			allowReason = fmt.Sprintf("allowed by rule %q", rule.ID)
		}
	}

	if matchedAllow {
		return true, allowReason
	}

	return false, "default deny: no matching allow rule found"
}

func (e *ABACPolicyEngine) matches(rule *PolicyRule, req AccessRequest) bool {
	// 1. Check Tenant
	if len(rule.Tenants) > 0 {
		tenantMatched := false
		for _, t := range rule.Tenants {
			if t == "*" || t == req.TenantID {
				tenantMatched = true
				break
			}
		}
		if !tenantMatched {
			return false
		}
	}

	// 2. Check Role
	if len(rule.Roles) > 0 {
		roleMatched := false
		for _, rRule := range rule.Roles {
			if rRule == "*" {
				roleMatched = true
				break
			}
			for _, rReq := range req.Roles {
				if rRule == rReq {
					roleMatched = true
					break
				}
			}
			if roleMatched {
				break
			}
		}
		if !roleMatched {
			return false
		}
	}

	// 3. Check Action
	actionMatched := false
	for _, a := range rule.Actions {
		if a == "*" || a == req.Action {
			actionMatched = true
			break
		}
	}
	if !actionMatched {
		return false
	}

	// 4. Check Resource
	resourceMatched := false
	for _, resPattern := range rule.Resources {
		if resPattern == "*" || resPattern == req.Resource {
			resourceMatched = true
			break
		}
		if strings.HasSuffix(resPattern, ":*") {
			prefix := strings.TrimSuffix(resPattern, ":*")
			if strings.HasPrefix(req.Resource, prefix+":") {
				resourceMatched = true
				break
			}
		}
	}
	if !resourceMatched {
		return false
	}

	return true
}

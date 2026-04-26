package core

import (
	"sync"
	"time"
)

// SLABreachRisk defines severity of upcoming deadline breach.
type SLABreachRisk string

const (
	SLARiskNone     SLABreachRisk = "NONE"
	SLARiskWarning  SLABreachRisk = "WARNING"
	SLARiskCritical SLABreachRisk = "CRITICAL"
	SLARiskBreached SLABreachRisk = "BREACHED"
)

// WorkflowSLAPolicy defines service level agreements for workflow run completion.
type WorkflowSLAPolicy struct {
	MaxDuration time.Duration
	WarningThresholdPct float64 // e.g. 75.0% of max duration
}

// SLATracker evaluates execution time against SLA deadline contracts.
type SLATracker struct {
	mu       sync.RWMutex
	policies map[string]WorkflowSLAPolicy // workflow name/type -> policy
}

// NewSLATracker creates an SLA tracker.
func NewSLATracker() *SLATracker {
	return &SLATracker{
		policies: make(map[string]WorkflowSLAPolicy),
	}
}

// SetPolicy sets or updates an SLA policy for a workflow type.
func (t *SLATracker) SetPolicy(workflowType string, policy WorkflowSLAPolicy) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if policy.WarningThresholdPct <= 0 {
		policy.WarningThresholdPct = 75.0
	}
	t.policies[workflowType] = policy
}

// EvaluateRisk checks elapsed duration and determines breach risk category.
func (t *SLATracker) EvaluateRisk(workflowType string, startTime time.Time, now time.Time) SLABreachRisk {
	t.mu.RLock()
	defer t.mu.RUnlock()

	policy, exists := t.policies[workflowType]
	if !exists || policy.MaxDuration <= 0 {
		return SLARiskNone
	}

	elapsed := now.Sub(startTime)
	if elapsed >= policy.MaxDuration {
		return SLARiskBreached
	}

	pctElapsed := (float64(elapsed) / float64(policy.MaxDuration)) * 100.0
	if pctElapsed >= 90.0 {
		return SLARiskCritical
	} else if pctElapsed >= policy.WarningThresholdPct {
		return SLARiskWarning
	}

	return SLARiskNone
}

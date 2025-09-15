package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// AdmissionDecision is the outcome of the admission control check.
type AdmissionDecision string

const (
	AdmissionGranted   AdmissionDecision = "granted"
	AdmissionThrottled AdmissionDecision = "throttled"
	AdmissionRejected  AdmissionDecision = "rejected"
)

// AdmissionRequest encapsulates metadata for a new workflow submission.
type AdmissionRequest struct {
	WorkflowID   string
	TenantID     string
	Priority     int
	EstimatedCPU float64 // normalized [0.0, 1.0]
	EstimatedMem float64 // normalized [0.0, 1.0]
	SubmittedAt  time.Time
}

// AdmissionResult is returned after evaluating an admission request.
type AdmissionResult struct {
	Request     AdmissionRequest
	Decision    AdmissionDecision
	Reason      string
	EvaluatedAt time.Time
}

// TenantAdmissionLimits defines rate and concurrency limits for a tenant.
type TenantAdmissionLimits struct {
	MaxConcurrentWorkflows int
	MaxSubmissionsPerMin   int
	MaxCPUShare            float64
	MaxMemShare            float64
}

// DefaultAdmissionLimits returns conservative limits for standard tenants.
func DefaultAdmissionLimits() TenantAdmissionLimits {
	return TenantAdmissionLimits{
		MaxConcurrentWorkflows: 50,
		MaxSubmissionsPerMin:   100,
		MaxCPUShare:            0.25,
		MaxMemShare:            0.30,
	}
}

// tenantAdmissionState tracks the current admission state for one tenant.
type tenantAdmissionState struct {
	Limits       TenantAdmissionLimits
	ActiveCount  int
	SubmitTimes  []time.Time // rolling minute window
	AllocatedCPU float64
	AllocatedMem float64
}

// AdmissionController evaluates incoming workflow submissions against per-tenant quotas.
type AdmissionController struct {
	mu      sync.Mutex
	tenants map[string]*tenantAdmissionState
}

// NewAdmissionController creates an empty admission controller.
func NewAdmissionController() *AdmissionController {
	return &AdmissionController{
		tenants: make(map[string]*tenantAdmissionState),
	}
}

// SetLimits configures or updates per-tenant admission limits.
func (a *AdmissionController) SetLimits(tenantID string, limits TenantAdmissionLimits) {
	a.mu.Lock()
	defer a.mu.Unlock()

	state, ok := a.tenants[tenantID]
	if !ok {
		state = &tenantAdmissionState{}
		a.tenants[tenantID] = state
	}
	state.Limits = limits
}

// Admit evaluates the admission request and returns a decision.
func (a *AdmissionController) Admit(req AdmissionRequest) *AdmissionResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := &AdmissionResult{
		Request:     req,
		EvaluatedAt: time.Now(),
	}

	state, ok := a.tenants[req.TenantID]
	if !ok {
		// Unknown tenant gets default limits
		state = &tenantAdmissionState{Limits: DefaultAdmissionLimits()}
		a.tenants[req.TenantID] = state
	}

	// Check concurrent workflow limit
	if state.ActiveCount >= state.Limits.MaxConcurrentWorkflows {
		result.Decision = AdmissionThrottled
		result.Reason = fmt.Sprintf("concurrent limit reached: %d/%d", state.ActiveCount, state.Limits.MaxConcurrentWorkflows)
		return result
	}

	// Check submissions-per-minute rate
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	fresh := state.SubmitTimes[:0]
	for _, t := range state.SubmitTimes {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	state.SubmitTimes = fresh

	if len(state.SubmitTimes) >= state.Limits.MaxSubmissionsPerMin {
		result.Decision = AdmissionThrottled
		result.Reason = fmt.Sprintf("submission rate exceeded: %d/%d per min", len(state.SubmitTimes), state.Limits.MaxSubmissionsPerMin)
		return result
	}

	// Check resource share
	if req.EstimatedCPU > 0 && state.AllocatedCPU+req.EstimatedCPU > state.Limits.MaxCPUShare {
		result.Decision = AdmissionRejected
		result.Reason = fmt.Sprintf("CPU share would exceed limit: %.2f+%.2f > %.2f", state.AllocatedCPU, req.EstimatedCPU, state.Limits.MaxCPUShare)
		return result
	}

	if req.EstimatedMem > 0 && state.AllocatedMem+req.EstimatedMem > state.Limits.MaxMemShare {
		result.Decision = AdmissionRejected
		result.Reason = fmt.Sprintf("memory share would exceed limit: %.2f+%.2f > %.2f", state.AllocatedMem, req.EstimatedMem, state.Limits.MaxMemShare)
		return result
	}

	// Admit
	state.ActiveCount++
	state.SubmitTimes = append(state.SubmitTimes, now)
	if req.EstimatedCPU > 0 {
		state.AllocatedCPU += req.EstimatedCPU
	}
	if req.EstimatedMem > 0 {
		state.AllocatedMem += req.EstimatedMem
	}

	result.Decision = AdmissionGranted
	result.Reason = "all checks passed"
	return result
}

// Release decrements the active count for a tenant (call on workflow completion).
func (a *AdmissionController) Release(tenantID string, cpu, mem float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	state, ok := a.tenants[tenantID]
	if !ok {
		return
	}

	if state.ActiveCount > 0 {
		state.ActiveCount--
	}
	state.AllocatedCPU -= cpu
	if state.AllocatedCPU < 0 {
		state.AllocatedCPU = 0
	}
	state.AllocatedMem -= mem
	if state.AllocatedMem < 0 {
		state.AllocatedMem = 0
	}
}

// Stats returns a human-readable summary of a tenant's admission state.
func (a *AdmissionController) Stats(tenantID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	state, ok := a.tenants[tenantID]
	if !ok {
		return fmt.Sprintf("tenant %s: unknown", tenantID)
	}
	return fmt.Sprintf("tenant=%s active=%d cpu=%.2f mem=%.2f submissions_1m=%d",
		tenantID, state.ActiveCount, state.AllocatedCPU, state.AllocatedMem, len(state.SubmitTimes))
}

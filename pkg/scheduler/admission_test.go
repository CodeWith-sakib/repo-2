package scheduler

import (
	"testing"
	"time"
)

func TestAdmissionController_GrantAndRelease(t *testing.T) {
	ac := NewAdmissionController()
	ac.SetLimits("tenant-a", TenantAdmissionLimits{
		MaxConcurrentWorkflows: 3,
		MaxSubmissionsPerMin:   100,
		MaxCPUShare:            0.5,
		MaxMemShare:            0.5,
	})

	req := AdmissionRequest{
		WorkflowID:   "wf-001",
		TenantID:     "tenant-a",
		EstimatedCPU: 0.1,
		EstimatedMem: 0.1,
		SubmittedAt:  time.Now(),
	}

	result := ac.Admit(req)
	if result.Decision != AdmissionGranted {
		t.Errorf("expected granted, got %s: %s", result.Decision, result.Reason)
	}

	ac.Release("tenant-a", 0.1, 0.1)

	stats := ac.Stats("tenant-a")
	if stats == "" {
		t.Error("stats should not be empty")
	}
}

func TestAdmissionController_ConcurrentLimit(t *testing.T) {
	ac := NewAdmissionController()
	ac.SetLimits("tenant-b", TenantAdmissionLimits{
		MaxConcurrentWorkflows: 2,
		MaxSubmissionsPerMin:   100,
		MaxCPUShare:            1.0,
		MaxMemShare:            1.0,
	})

	req := func(id string) AdmissionRequest {
		return AdmissionRequest{WorkflowID: id, TenantID: "tenant-b", SubmittedAt: time.Now()}
	}

	if r := ac.Admit(req("wf-1")); r.Decision != AdmissionGranted {
		t.Fatalf("wf-1 should be admitted: %s", r.Reason)
	}
	if r := ac.Admit(req("wf-2")); r.Decision != AdmissionGranted {
		t.Fatalf("wf-2 should be admitted: %s", r.Reason)
	}
	if r := ac.Admit(req("wf-3")); r.Decision != AdmissionThrottled {
		t.Errorf("wf-3 should be throttled at limit 2, got %s", r.Decision)
	}
}

func TestAdmissionController_CPUShareRejection(t *testing.T) {
	ac := NewAdmissionController()
	ac.SetLimits("tenant-c", TenantAdmissionLimits{
		MaxConcurrentWorkflows: 100,
		MaxSubmissionsPerMin:   100,
		MaxCPUShare:            0.5,
		MaxMemShare:            1.0,
	})

	r1 := ac.Admit(AdmissionRequest{WorkflowID: "wf-1", TenantID: "tenant-c", EstimatedCPU: 0.4, SubmittedAt: time.Now()})
	if r1.Decision != AdmissionGranted {
		t.Fatalf("first admission should pass: %s", r1.Reason)
	}

	r2 := ac.Admit(AdmissionRequest{WorkflowID: "wf-2", TenantID: "tenant-c", EstimatedCPU: 0.2, SubmittedAt: time.Now()})
	if r2.Decision != AdmissionRejected {
		t.Errorf("second admission should be rejected for CPU: got %s: %s", r2.Decision, r2.Reason)
	}
}

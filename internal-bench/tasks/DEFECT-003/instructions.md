# Benchmark Task: DEFECT-003 - Fair Queue Quota Off-By-One Enqueue Rejection

## Category
`off-by-one`

## Target Component
`pkg/scheduler` (target file: `pkg/scheduler/fair_queue_quota.go`)

## Symptom Description
Tenant queue enqueuer rejects new run submissions when queued count equals max allowed queued runs minus one.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/scheduler/fair_queue_quota.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

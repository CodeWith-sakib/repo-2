# Benchmark Task: DEFECT-007 - Worker Work-Stealing Stats Data Race

## Category
`race-condition`

## Target Component
`pkg/worker` (target file: `pkg/worker/priority_worker.go`)

## Symptom Description
Worker steal counter is incremented across peer threads without atomic operations or lock synchronization.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/worker/priority_worker.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

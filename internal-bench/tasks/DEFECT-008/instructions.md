# Benchmark Task: DEFECT-008 - Lock Ordering Inversion In Concurrency Governor

## Category
`deadlock`

## Target Component
`pkg/scheduler` (target file: `pkg/scheduler/throttler.go`)

## Symptom Description
Simultaneous rate limit token replenishment and concurrency slot acquisition acquire governor mutex and tenant mutex in opposite order.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/scheduler/throttler.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

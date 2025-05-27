# Benchmark Task: DEFECT-010 - Task Coordinator Cancels Root Context

## Category
`context-cancellation`

## Target Component
`pkg/worker` (target file: `pkg/worker/deadline.go`)

## Symptom Description
Task deadline timeout callback accidentally cancels the root parent context instead of only child step execution context.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/worker/deadline.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

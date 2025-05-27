# Benchmark Task: DEFECT-004 - Run State Direct Transition Pending to Completed

## Category
`state-machine-bypass`

## Target Component
`pkg/statemachine` (target file: `pkg/statemachine/state.go`)

## Symptom Description
Workflow run state machine fails to reject direct state mutation from PENDING directly to COMPLETED without going through RUNNING.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/statemachine/state.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

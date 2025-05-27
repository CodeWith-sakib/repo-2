# Benchmark Task: DEFECT-005 - Step Terminal Failed State Mutation Bypass

## Category
`state-machine-bypass`

## Target Component
`pkg/statemachine` (target file: `pkg/statemachine/step_state.go`)

## Symptom Description
Terminal step state FAILED permits transition back into RUNNING state violating terminal state immutability.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/statemachine/step_state.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

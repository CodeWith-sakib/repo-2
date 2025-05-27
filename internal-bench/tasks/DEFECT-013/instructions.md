# Benchmark Task: DEFECT-013 - Macro Evaluator Division Zero Error Suppression

## Category
`error-masking`

## Target Component
`pkg/core` (target file: `pkg/core/macro_eval.go`)

## Symptom Description
Macro expression evaluator returns 0 instead of returning error when dividing by zero in arithmetic operations.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/core/macro_eval.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

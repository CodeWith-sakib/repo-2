# Benchmark Task: DEFECT-012 - Database Circuit Breaker Error Swallowing

## Category
`error-masking`

## Target Component
`pkg/storage/postgres` (target file: `pkg/storage/postgres/circuit.go`)

## Symptom Description
Circuit breaker probe execution ignores underlying database driver network error and returns nil, masking failures.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/storage/postgres/circuit.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

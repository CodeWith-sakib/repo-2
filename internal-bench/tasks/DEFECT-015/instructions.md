# Benchmark Task: DEFECT-015 - Postgres Pool Monitor Goroutine Leak

## Category
`resource-leak`

## Target Component
`pkg/storage/postgres` (target file: `pkg/storage/postgres/pool.go`)

## Symptom Description
Postgres pool monitor health loop does not terminate upon pool close, leaving a perpetual background goroutine.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/storage/postgres/pool.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

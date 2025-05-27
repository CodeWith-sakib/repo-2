# Benchmark Task: DEFECT-009 - Savepoint Rollback Lock Inversion

## Category
`deadlock`

## Target Component
`pkg/storage/memory` (target file: `pkg/storage/memory/transaction.go`)

## Symptom Description
Nested transaction commit and concurrent rollback lock parent and child snapshot entries in contradictory order under high concurrency.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/storage/memory/transaction.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

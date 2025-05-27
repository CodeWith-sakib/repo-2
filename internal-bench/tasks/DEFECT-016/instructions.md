# Benchmark Task: DEFECT-016 - Burst Capacity Duration Int32 Truncation Overflow

## Category
`integer-overflow`

## Target Component
`pkg/core` (target file: `pkg/core/rate_limiter.go`)

## Symptom Description
Rate limiter converts large burst nanoseconds using int32 cast instead of int64, causing negative duration on large burst limits.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/core/rate_limiter.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

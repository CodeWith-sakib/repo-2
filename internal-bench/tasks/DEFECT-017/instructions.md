# Benchmark Task: DEFECT-017 - Percentile Histogram Bin Uint32 Overflow

## Category
`integer-overflow`

## Target Component
`pkg/metrics` (target file: `pkg/metrics/histogram.go`)

## Symptom Description
Sample count accumulation overflows uint32 bin index when high-frequency traffic exceeds 4.29 billion events.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/metrics/histogram.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

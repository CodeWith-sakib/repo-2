# Benchmark Task: DEFECT-002 - Cache Capacity Premature Eviction

## Category
`off-by-one`

## Target Component
`pkg/cache` (target file: `pkg/cache/cache.go`)

## Symptom Description
Cache evicts the least recently used element when item count is strictly capacity minus one instead of reaching capacity.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/cache/cache.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

# Benchmark Task: DEFECT-022 - Negative Span Duration Calculation In Tracer

## Category
`metric-skew`

## Target Component
`pkg/metrics` (target file: `pkg/metrics/otel.go`)

## Symptom Description
Tracer endFn subtracts EndTime from StartTime instead of StartTime from EndTime, reporting negative span latency.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/metrics/otel.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

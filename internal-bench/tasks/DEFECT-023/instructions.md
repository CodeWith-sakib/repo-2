# Benchmark Task: DEFECT-023 - Queue High Watermark Undercount

## Category
`metric-skew`

## Target Component
`pkg/scheduler` (target file: `pkg/scheduler/priority_queue_metrics.go`)

## Symptom Description
Priority queue telemetry records queue depth before enqueuing element, failing to observe maximum peak queue length.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/scheduler/priority_queue_metrics.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

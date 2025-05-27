# Benchmark Task: DEFECT-020 - CloudEvents Serialization Subject Drop

## Category
`serialization-loss`

## Target Component
`pkg/events` (target file: `pkg/events/serialization_advanced.go`)

## Symptom Description
CloudEvents v1.0 JSON marshaler forgets to serialize the Subject attribute into the wire payload.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/events/serialization_advanced.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

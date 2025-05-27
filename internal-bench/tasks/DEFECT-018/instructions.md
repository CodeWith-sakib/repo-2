# Benchmark Task: DEFECT-018 - Environment Variable Overrides CLI Flag

## Category
`config-precedence`

## Target Component
`pkg/config` (target file: `pkg/config/config.go`)

## Symptom Description
Configuration precedence incorrectly applies environment variable overrides after CLI flag evaluation.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/config/config.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

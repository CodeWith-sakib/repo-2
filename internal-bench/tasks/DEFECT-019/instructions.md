# Benchmark Task: DEFECT-019 - DeepMerge Nested Overlay Key Deletion

## Category
`config-precedence`

## Target Component
`pkg/config` (target file: `pkg/config/merger.go`)

## Symptom Description
DeepMerge algorithm overwrites entire nested map rather than recursively merging child keys.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/config/merger.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

# Benchmark Task: DEFECT-025 - Disconnected Circular Subgraph False Negative

## Category
`cycle-detection`

## Target Component
`pkg/core` (target file: `pkg/core/dag.go`)

## Symptom Description
Kahn algorithm topological sort exits early when root nodes are processed, missing disconnected circular clusters.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/core/dag.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

# Benchmark Task: DEFECT-024 - Tarjan SCC Detector Self-Loop False Negative

## Category
`cycle-detection`

## Target Component
`pkg/core` (target file: `pkg/core/cycle_tarjan.go`)

## Symptom Description
Tarjan SCC algorithm ignores single-node cycles where a step lists itself in DependsOn.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/core/cycle_tarjan.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

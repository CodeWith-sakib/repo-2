# Benchmark Task: DEFECT-006 - Unsynchronized Map Access In In-Memory Task Queue Ack

## Category
`race-condition`

## Target Component
`pkg/storage/memory` (target file: `pkg/storage/memory/queue.go`)

## Symptom Description
Task queue Ack method deletes from the active lease map without holding the internal write mutex, tripping Go race detector.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/storage/memory/queue.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

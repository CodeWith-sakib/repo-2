# Benchmark Task: DEFECT-014 - Event Deduplicator Background Ticker Leak

## Category
`resource-leak`

## Target Component
`pkg/events` (target file: `pkg/events/dedup.go`)

## Symptom Description
Event deduplicator Close() method closes stopCh but fails to call Stop() on the time.Ticker, leaking runtime timers.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/events/dedup.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

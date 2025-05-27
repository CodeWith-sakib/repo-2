# Benchmark Task: DEFECT-011 - HTTP Middleware Drops Upstream Context Cancellation

## Category
`context-cancellation`

## Target Component
`pkg/api/http` (target file: `pkg/api/http/middleware.go`)

## Symptom Description
Request context wrapper creates disconnected background context, preventing worker cancellation when client disconnects.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/api/http/middleware.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

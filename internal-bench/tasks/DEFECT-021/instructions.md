# Benchmark Task: DEFECT-021 - gRPC Wire Codec Buffer Boundary Payload Truncation

## Category
`serialization-loss`

## Target Component
`pkg/api/grpc` (target file: `pkg/api/grpc/codec.go`)

## Symptom Description
Custom varint wire decoder miscalculates multi-chunk payload boundary length, truncating binary trailing bytes.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/api/grpc/codec.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

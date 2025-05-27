# Benchmark Task: DEFECT-001 - Macro Substring Truncation On Max Length

## Category
`off-by-one`

## Target Component
`pkg/core` (target file: `pkg/core/template_macro.go`)

## Symptom Description
When evaluating substring macros or path slicers in template parameters, extracting substring with length equal to slice capacity prematurely truncates the final rune.

## Instructions for Agent
1. Reproduce the defect using the targeted test suite.
2. Locate the root cause in `pkg/core/template_macro.go`.
3. Apply a clean, minimal fix preserving API compatibility and architectural integrity.
4. Verify that all package unit tests and race checks pass (`make test-race`).

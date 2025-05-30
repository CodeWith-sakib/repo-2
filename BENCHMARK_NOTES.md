# Benchmark Notes

This directory contains the ground-truth benchmark suite for AI coding agent evaluations.

## Structure
- `defects.yaml`: Master manifest cataloging 25 verified defects across 12 requirement categories.
- `tasks/<defect-id>/`: Sand-style benchmark task bundle containing:
  - `instructions.md`: Unbiased symptom description and agent instructions.
  - `solution.patch`: Minimal canonical fix patch.
  - `evidence/`: Execution logs demonstrating F2P (Fail-to-Pass) and P2P (Pass-to-Pass) reproducibility.

## Defect Categories
1. Off-by-one / boundary conditions
2. State machine invalid transition bypass
3. Race condition / concurrent map access
4. Deadlock in lock acquisition order
5. Context cancellation / leak
6. Error masking / wrapped error loss
7. Resource leak (goroutine, ticker)
8. Integer overflow / conversion truncation
9. Configuration precedence inversion
10. Serialization / round-trip field loss
11. Metric / telemetry undercount / skew
12. Topological cycle detection false negative

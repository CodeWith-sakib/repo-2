# KestrelFlow Industrial Enhancement Plan

Based on the baseline audit in `AUDIT.md`, this plan details the systematic closure of the remaining production source code gap toward the 32,000–40,000 LOC target while enforcing the per-commit verification gate.

---

## 1. Target Objectives & Gap Analysis

- **Current Production LOC**: 10,610 LOC
- **Target Production LOC**: 32,000–40,000 LOC (Gap: ~21,400–29,400 LOC)
- **Commit History**: 155 commits (target ≥ 150 satisfied; will continue incremental additions with per-commit verification gate)
- **Per-Commit Verification Gate**:
  1. Formatter: `gofmt -l .` must report no changes
  2. Full Build: `make build` (`/opt/homebrew/bin/go build -buildvcs=false ./...`) must pass cleanly
  3. Static Analysis / Linter: `make vet` and `make staticcheck` must report 0 warnings
  4. Unit / Package Tests: Touched package tests and dependents must pass
  5. Concurrency / Race Check: `go test -race` on concurrency-relevant packages
  6. Diff Size Sanity: No commit > 8% total LOC
  7. Commit Trailer: Every commit message must include the gate trailer:
     `Gate: build=pass lint=pass tests=pass race=pass`

---

## 2. Planned Subsystem Expansions (Closing the LOC Gap)

We will expand the existing 16 subsystems with deep, domain-authentic, production-grade Go implementations and accompanying tests:

### Phase 2A — Core Engine & Expression System
- **Rich AST Visitor & Rewriter**: Full AST node walker, dead-code pruning, constant folding tree, boolean normalizer (CNF/DNF).
- **Advanced JSONPath & Expression Evaluator**: Recursive descent JSONPath slice evaluator, function registry, aggregation operators (sum, avg, min, max, stddev).
- **Workflow Macro & Template Interpolator**: Multi-pass template variable resolution, default values, filter pipes.
- **Topological Graph Algorithms**: Tarjan articulation point finder, biconnected components, transitive reduction, critical path schedule optimizer.

### Phase 2B — Storage & Persistence Engines
- **PostgreSQL Advanced AST Query Engine**: Window functions (`ROW_NUMBER`, `RANK`), CTE (`WITH RECURSIVE`), subquery aliasing, table join builders.
- **WAL & Storage Replication Mock**: Append-only write-ahead log with segment roll-over, checksum verification, log replay coordinator.
- **Distributed Lease Fencing & Lock Manager**: Monotonic fencing epoch coordinator, leader election simulator, lease heartbeat reclaimer.

### Phase 2C — Scheduler, Worker & Concurrency Layer
- **Fair-Share Priority Scheduler**: Multi-queue virtual time scheduler, tenant deficit round-robin (DRR), starvation prevention.
- **Worker Affinity & Resource Constraint Matcher**: CPU/memory/GPU slot allocation, label selectors, taint/toleration admission controller.
- **Graceful Drain & Task Handoff Coordinator**: Two-phase drain, flight tracker, peer task migration protocol.

### Phase 2D — API, Transport & Security Layer
- **OpenAPI v3 & JSON Schema Generator**: Runtime struct inspection, OpenAPI 3.0 schema generation, request/response validation filters.
- **gRPC Stream Multiplexer & Binary Wire Protocol**: Varint framed chunking, flow control window tracker, backpressure buffer.
- **Cryptographic Audit & Key Management**: Merkle tree audit log verifier, key rotation manager, mTLS certificate validator.

### Phase 2E — Observability, Telemetry & Maintenance
- **Prometheus Metric Exposition & HDR Latency Histogram**: Quantile estimator, counter/gauge aggregators, text exposition format serializer.
- **OpenTelemetry Trace Exporter**: Distributed trace context propagation (W3C TraceContext headers), span processor, batch exporter.
- **Background Maintenance & Compaction Engine**: Partition compactor, soft-deleted record purger, tombstone vacuum.

---

## 3. Defect Category Distribution & Verification

All 12 categories are maintained with 25 verified defects in `internal-bench/defects.yaml` and Sand task packages in `internal-bench/tasks/DEFECT-001` through `DEFECT-025`.

---

## 4. Commit Log

| Commit | Phase | Gate | Summary |
|---|---|---|---|

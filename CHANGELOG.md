# Changelog

All notable changes to KestrelFlow will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-golden] - 2026-05-20

### Added
- **Subsystems Architecture**: 16 fully implemented first-party subsystems.
- **Core Engine**: DAG construction, Kahn & Tarjan cycle detection, AST expressions, JSON Schema validator, cron parser, and macro expansion.
- **State Machine**: Deterministic transition matrix for workflows and steps with Saga compensation engine.
- **Scheduler**: Priority-tiered fair queue, tenant concurrency governor, worker affinity matcher, and load balancer.
- **Worker Pool**: Resilient worker pool with heartbeats, work-stealing priority worker, and graceful drain coordinator.
- **Storage**: Multi-backend engine supporting In-Memory MVCC snapshot store and PostgreSQL with AST query builder, savepoints, and circuit breaker.
- **API Surfaces**: Dual HTTP REST (RFC 7807 problem details) and gRPC wire services with reflection struct validator.
- **Operator CLI**: Subcommands for server, worker, workflow, and run lifecycle management.
- **Config**: 4-tier precedence resolver (Flag > Env > File > Default) with dynamic file watcher and hot-reloading.
- **Plugins**: Task plugin interfaces with first-party HTTP, Shell, SQL, and Transform plugins, protected by process sandbox.
- **Observability**: Prometheus metrics exposition, sliding-window HDR quantile histogram, OpenTelemetry tracer, and server-rendered HTML dashboard.
- **Fuzzing & Property Tests**: Comprehensive native fuzz targets and quickcheck invariant property tests.
- **Benchmark Suite**: 25 Sand-style packaged defect verification benchmarks.

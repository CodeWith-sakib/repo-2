# KestrelFlow Design Document

## 1. Product Overview & Vision

**KestrelFlow** is an embeddable and distributed workflow-orchestration engine built for platform engineering teams. It coordinates complex, multi-stage Directed Acyclic Graph (DAG) pipelines with durable execution guarantees, granular state-machine transitions, automatic retries with configurable backoff policies, tenant-aware fair scheduling, and zero external runtime dependencies when running in self-contained mode.

### Key Tenets
1. **Durable & Deterministic Execution**: Workflows transition through strict, validated state machines. Execution history is journaled to durable storage (PostgreSQL in production, in-memory with identical semantics for testing and embedded usage).
2. **Pluggable Extensibility**: Task execution is abstracted behind a clean plugin contract (`TaskHandler`), enabling platform teams to integrate HTTP endpoints, shell scripts, container jobs, or custom enterprise workers.
3. **Observability by Design**: Metrics, structured logs, an event bus, webhook notifications, and a lightweight server-rendered status dashboard provide real-time inspection of pipeline health.
4. **Resilient Concurrency**: Workers operate on a lease-based heartbeat mechanism with graceful drain on SIGINT/SIGTERM, dynamic goroutine pooling, and context-propagated cancellation.

---

## 2. Subsystem Package Architecture

All 16 required subsystems are assigned explicit, modular package boundaries within the repository:

| Subsystem # | Subsystem Name | Package Path | Primary Responsibility |
|---|---|---|---|
| 1 | Workflow Engine & State Machine | `pkg/core`, `pkg/statemachine` | Workflow/step DAG definitions, dependency validation, topological sort, deterministic state transitions |
| 2 | Scheduler & Worker Pool | `pkg/scheduler`, `pkg/worker` | Fair task queueing, priority dispatch, goroutine worker pool, lease heartbeats, cancellation cascading |
| 3 | Persistence Layer | `pkg/storage`, `pkg/storage/postgres`, `pkg/storage/memory` | Durable storage abstraction, PostgreSQL driver implementation, thread-safe in-memory store for fast tests |
| 4 | Hot Definition Cache | `pkg/cache` | LRU and TTL in-memory cache for parsed workflow schemas with explicit event-driven invalidation |
| 5 | HTTP REST API | `pkg/api/http`, `pkg/api/types` | Versioned RESTful endpoints (`/api/v1/...`) for workflow creation, execution management, run inspection |
| 6 | gRPC API Surface | `pkg/api/grpc` | High-performance RPC interface mirroring HTTP operations for low-latency cluster inter-service calls |
| 7 | Operator CLI | `pkg/cli`, `cmd/kestrel` | Full-featured command-line interface for operators (`server`, `worker`, `workflow`, `run`, `admin`) |
| 8 | Plugin Interface & Handlers | `pkg/plugin`, `pkg/plugin/builtin/http`, `pkg/plugin/builtin/shell` | Extensible `TaskHandler` contract and built-in executors (HTTP Webhook/API caller, Script/Shell runner) |
| 9 | Retry, Backoff & Contexts | `pkg/retry` | Exponential, linear, and jittered backoff strategies, retry policies, context propagation and deadline management |
| 10 | Events & Webhook Dispatch | `pkg/events`, `pkg/webhook` | Internal pub/sub event bus, webhook delivery engine with exponential retry, HMAC signatures, dead-letter storage |
| 11 | Metrics & Observability | `pkg/metrics`, `pkg/debug` | Telemetry counters, duration histograms, Prometheus-compatible `/metrics` endpoint, pprof `/debug` handlers |
| 12 | Configuration Engine | `pkg/config` | 4-tier configuration loader: CLI Flags > Environment Variables > Config File > Defaults |
| 13 | Auth & RBAC Middleware | `pkg/auth` | Bearer token authentication and role-based access control (`Viewer`, `Operator`, `Admin`) |
| 14 | Background Maintenance | `pkg/maintenance` | Background workers for run history retention expiry, log compaction, orphaned lease cleanup |
| 15 | Status Dashboard | `pkg/dashboard` | Server-rendered HTML dashboard using Go `html/template` with embedded CSS for live visual inspection |
| 16 | Schema Migrations & Fixtures | `pkg/storage/migrations`, `fixtures` | Up/down SQL schema migration engine and representative pipeline fixtures (ETL, fan-out, approvals) |

---

## 3. Core Data Model & State Machine

### 3.1 Entity Model
- **WorkflowDefinition**: Immutable schema containing identifier, version, name, description, parameters schema, timeout, and a list of `StepDefinition` nodes forming a DAG.
- **StepDefinition**: Single unit of execution with unique step ID, task type, handler configuration, prerequisites (`DependsOn`), retry policy, and timeout.
- **WorkflowRun**: Execution instance of a workflow, holding run ID, workflow ID/version, input parameters, overall state (`Pending`, `Running`, `Suspended`, `Completed`, `Failed`, `Cancelled`), timestamps, and error summary.
- **StepRun**: Execution record of a single step within a run, tracking attempt count, state (`Pending`, `Queued`, `Running`, `Completed`, `Failed`, `Skipped`), input/output payloads, worker lease ID, and started/finished timestamps.
- **Lease**: Distributed lock acquired by a worker on a step run, renewed via periodic heartbeats. If a lease expires, the scheduler reassigns the step.

### 3.2 State Transition Invariants
- `Pending` -> `Running` -> `Completed` | `Failed` | `Cancelled` | `Suspended`
- `Suspended` -> `Running` | `Cancelled`
- Terminal states (`Completed`, `Failed`, `Cancelled`) are immutable.

---

## 4. Concurrency & Reliability Architecture

1. **Scheduler Loop**: Ticks on a configurable interval or event notification, evaluating runnable steps whose upstream dependencies have succeeded.
2. **Worker Pool**: Maintains bounded goroutine worker pool with work-stealing / fair-dispatch queue. Context cancellation cascades down from the workflow run to active step goroutines.
3. **Lease Heartbeats**: Active workers run background heartbeat goroutines to keep step leases alive. A lease expiry detector marks orphaned tasks as failed or ready for re-queue.
4. **Deterministic Testing**: Persistence interfaces define identical transactional semantics for both PostgreSQL and in-memory implementations, enabling high-speed `-race` verification with zero network overhead.

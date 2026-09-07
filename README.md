# KestrelFlow

[![CI](https://github.com/CodeWith-sakib/repo-2/actions/workflows/ci.yml/badge.svg)](https://github.com/CodeWith-sakib/repo-2/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://golang.org)

**KestrelFlow** is a high-performance, distributed, durable workflow-orchestration engine written in Go. It provides transactional state guarantees, fair-share scheduling, dynamic worker pools with heartbeat-based lease fencing, pluggable task handlers, and first-party operational observability for complex Directed Acyclic Graph (DAG) pipelines.

---

## Highlights & Capabilities

- **Durable DAG Execution**: Deterministic finite-state machine governing workflow and step lifecycles with strict legal transition validation and terminal-state immutability.
- **Fair-Share Scheduler & Worker Pool**: Priority-aware scheduling, deficit round-robin queueing, worker capability scoring, adaptive concurrency limits, and lease heartbeats with auto-reclaim.
- **Dual Persistence Backends**: Production PostgreSQL backend utilizing parameterized queries alongside an in-memory transactional store sharing identical semantics for fast, zero-dependency testing.
- **Pluggable Task Handlers**: Extensible `TaskHandler` interface with built-in handlers for HTTP endpoints, sandboxed Shell commands with environment sanitization, SQL queries, and JSONPath transforms.
- **Operator Dashboard & CLI**: Server-rendered HTML status dashboard template and a rich command-line tool (`kestrel`) for cluster administration, workflow submissions, and run inspection.
- **Resilient Webhook & Event Bus**: Asynchronous pub/sub event bus supporting CNCF CloudEvents v1.0, HMAC-SHA256 signature verification, exponential retry with jitter, and dead-letter queue routing.
- **Full-Stack Observability**: Built-in Prometheus-compatible metrics exposition, W3C traceparent header extraction, structured logging, and HTTP `/healthz` monitoring.

---

## Architecture Overview

```
                        +----------------------------+
                        |  CLI (`kestrel`) / HTTP    |
                        +--------------+-------------+
                                       |
                                       v
                        +----------------------------+
                        |  REST API & Middleware     |
                        |  - ProblemDetails RFC 7807 |
                        |  - Panic Recovery Handler  |
                        +--------------+-------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
    +------------------------------+       +------------------------------+
    |     Scheduler & Queues       |       |       DAG State Machine      |
    |  - Fair-Share DRR Queue      |       |  - Topological Sort (Kahn)   |
    |  - Priority Lease Allocator  |       |  - Transition Invariants     |
    +--------------+---------------+       +--------------+---------------+
                   |                                       |
                   v                                       v
    +------------------------------+       +------------------------------+
    |      Worker Pool             |       |      Persistence Layer       |
    |  - Heartbeat Lease Extender  | <---> |  - PostgreSQL Store          |
    |  - Task Execution Sandboxes  |       |  - In-Memory MVCC Store      |
    +--------------+---------------+       +------------------------------+
                   |
                   v
    +------------------------------+
    | Pluggable Handlers (HTTP/Cmd)|
    +------------------------------+
```

---

## Project Layout

```
cmd/
  kestrel/            Application main entrypoint and CLI runner
pkg/
  api/
    http/             RESTful HTTP API routes, middleware, and handlers
    grpc/             High-performance gRPC wire services
    types/            API data transfer objects and validation schemas
  cli/                CLI command definitions (server, worker, workflow, run)
  config/             4-tier configuration loader (Flags > Env > File > Defaults)
  core/               Workflow DAG primitives, validation, expression evaluator
  events/             Event bus, CloudEvents v1.0, binary journal, topic routing
  metrics/            Prometheus metrics registry and telemetry collectors
  plugin/             TaskHandler contracts, execution sandboxes, builtin handlers
  retry/              Backoff algorithms (exponential with jitter) and circuit breakers
  scheduler/          Task queueing, lease management, worker load balancers
  statemachine/       Deterministic state transition engines and Saga compensation
  storage/
    memory/           Thread-safe in-memory store with secondary indexes
    postgres/         PostgreSQL adapter, query builder, and migration engine
  webhook/            HMAC-signed webhook dispatchers and delivery queues
  worker/             Worker pools, execution supervision, and lease renewal
internal-bench/       Benchmark evaluation tasks and test harnesses
```

---

## Getting Started

### Prerequisites

- **Go**: 1.26 or higher
- **Make**: Standard build automation
- **Docker** (optional): For containerized execution

### Building from Source

```bash
# Build all packages and binaries
make build

# The kestrel binary will be available via Go build cache or ./cmd/kestrel
go build -o bin/kestrel ./cmd/kestrel
```

### Running KestrelFlow

Start the KestrelFlow server with in-memory storage (zero external dependencies):

```bash
./bin/kestrel server -host=127.0.0.1 -port=8080
```

Start a standalone worker process:

```bash
./bin/kestrel worker -concurrency=8
```

Inspect CLI usage and available subcommands:

```bash
./bin/kestrel --help
```

---

## Configuration

KestrelFlow resolves configuration through a 4-tier hierarchy:
`CLI Flags > Environment Variables > Configuration File > Defaults`

Copy the sample environment file to configure your local deployment:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|---|---|---|
| `KESTREL_CONFIG` | Path to optional JSON configuration file | *None* |
| `KESTREL_SERVER_HOST` | Host interface to bind HTTP listener | `127.0.0.1` |
| `KESTREL_SERVER_PORT` | Port for HTTP REST API | `8080` |
| `KESTREL_DB_TYPE` | Storage backend (`memory` or `postgres`) | `memory` |
| `KESTREL_DB_DSN` | PostgreSQL connection string | *None* |
| `KESTREL_WORKER_CONCURRENCY` | Concurrent step execution goroutines | `4` |

---

## HTTP REST API

The server exposes standard REST endpoints for managing workflow lifecycles:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness and readiness health probe |
| `POST` | `/api/v1/workflows` | Submit and register a workflow DAG definition |
| `GET` | `/api/v1/workflows/{id}` | Retrieve a workflow definition by identifier |
| `POST` | `/api/v1/runs` | Trigger execution of a registered workflow |
| `GET` | `/api/v1/runs/{id}` | Fetch status and step execution details for a run |

---

## Running with Docker

Build the multi-stage production container image:

```bash
docker build -t kestrelflow:latest .
```

Run the containerized server:

```bash
docker run --rm -p 8080:8080 --name kestrelflow kestrelflow:latest
```

Verify service health:

```bash
curl http://localhost:8080/healthz
```

---

## Testing & Quality Gates

KestrelFlow enforces complete test-suite verification:

```bash
# Run unit and package tests
make test

# Run tests with the Go race detector enabled
make test-race

# Run static analysis and linter checks
make vet
make staticcheck
```

---

## License

This project is licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

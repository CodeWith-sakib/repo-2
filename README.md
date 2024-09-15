# KestrelFlow

[![CI](https://github.com/kestrelflow/kestrelflow/actions/workflows/ci.yml/badge.svg)](https://github.com/kestrelflow/kestrelflow/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**KestrelFlow** is a distributed, durable workflow-orchestration engine built in Go. It enables platform teams to author, schedule, execute, and monitor resilient Directed Acyclic Graph (DAG) pipelines with transactional state guarantees, dynamic worker pools, pluggable task execution, and comprehensive operational observability.

## Highlights

- **Durable DAG State Engine**: Deterministic state machine tracking workflow runs and individual step states.
- **Fair-Share Scheduler & Worker Pool**: Priority-aware scheduling with lease-based heartbeat locks and graceful worker drains.
- **Dual Persistence Backends**: Native PostgreSQL support via `database/sql` alongside an in-memory repository providing identical transactional semantics for tests.
- **Pluggable Execution**: First-party task executors for HTTP APIs and Sandboxed Shell scripts, with an open `TaskHandler` plugin interface.
- **Operator Dashboard & CLI**: Server-rendered real-time HTML status dashboard and a rich command-line tool (`kestrel`).
- **Resilient Webhooks**: Event bus with signature-authenticated (HMAC-SHA256) webhook notifications, backoff retries, and dead-letter queues.
- **Observability**: Built-in Prometheus-compatible telemetry, request logging, and runtime debug endpoints.

## Getting Started

### Prerequisites
- Go 1.26 or later

### Building
```bash
make build
```

### Running Tests
```bash
make test
make test-race
```

## Architecture
See [DESIGN.md](DESIGN.md) for full architectural documentation, subsystem boundaries, and state transition specifications.

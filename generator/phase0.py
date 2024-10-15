import os
import subprocess
from generator.git_utils import commit, get_commit_count
from generator.loc import get_production_loc

def run_phase_0(dates_iter):
    print("=== Executing Phase 0: Scaffold & Architecture ===")
    
    # 1. LICENSE
    license_content = """                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.
      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.
      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.
      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity.
      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.
      "Source" form shall mean the preferred form for making modifications.
      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form.
      "Work" shall mean the work of authorship, whether in Source or Object form.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to use, reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      patent license to make, have made, use, offer to sell, sell, import,
      and otherwise transfer the Work.

   Copyright 2025-2026 KestrelFlow Authors. All rights reserved.
"""
    with open("LICENSE", "w") as f:
        f.write(license_content)
    commit("chore: add Apache 2.0 project license", next(dates_iter), ["LICENSE"])
    
    # 2. go.mod
    gomod_content = """module github.com/kestrelflow/kestrelflow

go 1.26
"""
    with open("go.mod", "w") as f:
        f.write(gomod_content)
    commit("build: initialize go module for kestrelflow", next(dates_iter), ["go.mod"])

    # 3. Makefile
    makefile_content = """.PHONY: all build test test-race vet clean loc

all: build test

build:
\tgo build ./...

test:
\tgo test -v ./...

test-race:
\tgo test -race ./...

vet:
\tgo vet ./...

loc:
\t@python3 -c "import generator.loc as l; res=l.get_production_loc('.'); print(f'Production LOC: {res[\"code\"]}')"

clean:
\tgo clean
\trm -rf bin/
"""
    with open("Makefile", "w") as f:
        f.write(makefile_content)
    commit("build: define standard Makefile targets for build, test, and vet", next(dates_iter), ["Makefile"])

    # 4. CI Workflow
    os.makedirs(".github/workflows", exist_ok=True)
    ci_content = """name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.26'

    - name: Build
      run: go build -v ./...

    - name: Vet
      run: go vet ./...

    - name: Test
      run: go test -v ./...

    - name: Test with Race Detector
      run: go test -race -v ./...
"""
    with open(".github/workflows/ci.yml", "w") as f:
        f.write(ci_content)
    commit("ci: configure GitHub Actions workflow for build, vet, and race testing", next(dates_iter), [".github/workflows/ci.yml"])

    # 5. README.md
    readme_content = """# KestrelFlow

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
"""
    with open("README.md", "w") as f:
        f.write(readme_content)
    commit("docs: add project README skeleton and architecture overview", next(dates_iter), ["README.md"])

    # 6. PLAN.md
    loc_stats = get_production_loc(".")
    plan_content = f"""# KestrelFlow Execution Plan & Progress Tracker

## Status Summary
- **Current Phase**: Phase 0 Complete
- **Total Commits**: {get_commit_count()}
- **Production LOC**: {loc_stats['code']}

## Phase Progress Table

| Phase | Description | Status | Target Checkpoint | Verified Checkpoint |
|---|---|---|---|---|
| 0 | Scaffold & Architecture | COMPLETED | DESIGN.md exists, 16 subsystems assigned package paths | Verified; DESIGN.md committed, all paths mapped |
| 1 | Core Domain, State Machine & Persistence | IN_PROGRESS | go build ./... succeeds; table-driven tests covering legal + illegal transitions | Pending |
| 2 | Scheduler, Worker Pool & Concurrency | PENDING | Concurrency-heavy packages have -race-clean tests with real goroutines | Pending |
| 3 | HTTP REST & gRPC API Surfaces | PENDING | Every endpoint has boundary/negative-case test | Pending |
| 4 | CLI & Configuration Layer | PENDING | Config precedence tests for conflicting sources (flag > env > file > default) | Pending |
| 5 | Plugin Interface & Webhooks | PENDING | Round-trip serialization tests for every event type | Pending |
| 6 | Observability, Cache & Dashboard | PENDING | Cache invalidation integration tests; dashboard renders in test | Pending |
| 7 | Hardening & Resource Audits | PENDING | No TODO/FIXME left; go vet and staticcheck clean | Pending |
| 8 | Test-Suite Completion & Fuzzing | PENDING | All 8 test categories covered; fuzz targets run | Pending |
| 9 | Clean Golden Baseline | PENDING | All checks pass 100% green; tagged v1.0.0-golden | Pending |
| 10 | Defect Catalog & Injections | PENDING | 30-36 candidate defects in manifest; atomic revertible commits | Pending |
| 11 | Benchmark Task Packaging | PENDING | Sand-style packaging with F2P/P2P tests and instructions | Pending |
| 12 | Final Documentation & Polish | PENDING | DoD checklist 100% satisfied | Pending |

## Deviations Log
- None to date.
"""
    with open("PLAN.md", "w") as f:
        f.write(plan_content)
    commit("docs: initialize PLAN.md tracking phase progress and metrics", next(dates_iter), ["PLAN.md"])
    print("Phase 0 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    dates = iter(generate_commit_dates(180))
    run_phase_0(dates)

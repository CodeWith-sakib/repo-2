# Contributing to KestrelFlow

Thank you for your interest in contributing to KestrelFlow!

## Prerequisites
- Go 1.25+ or Go 1.26+
- Make
- Git

## Development Workflow
1. **Format & Lint**: Ensure all code adheres to standard Go idioms:
   ```bash
   make vet
   make staticcheck
   ```
2. **Testing**: Run unit tests and race condition checks:
   ```bash
   make test
   make test-race
   ```
3. **Commit Standards**: Write clear, descriptive, imperative commit messages (e.g. `core: implement topological cycle detection`).

## Architecture Standards
- Zero external network dependencies in tests (use mock round-trippers or loopbacks).
- Strict race cleanliness: no un-synchronized map reads or pointer sharing across goroutines.

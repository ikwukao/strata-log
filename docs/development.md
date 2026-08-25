# Strata-Log Development Guide

This document describes the development workflow for contributing to Strata-Log.

---

## Prerequisites

Install:

* Go 1.26+
* Git
* Docker
* Docker Compose

Verify Go:

```bash
go version
```

Verify Docker:

```bash
docker version
```

Verify Docker Compose:

```bash
docker compose version
```

---

## Clone the Repository

```bash
git clone https://github.com/ikwukao/strata-log.git
cd strata-log
```

---

## Project Structure

```text
strata-log/
├── benchmarks/
│   ├── ingest_test.go
│   └── storage_test.go
├── cmd/
│   └── strata-log/
├── deployments/
├── docs/
├── internal/
│   ├── batcher/
│   ├── buffer/
│   ├── config/
│   ├── ingest/
│   ├── pipeline/
│   ├── query/
│   ├── resilience/
│   ├── storage/
│   └── telemetry/
├── tests/
├── CONTRIBUTING.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── README.md
├── ROADMAP.md
├── go.mod
└── go.sum
```

---

## Development Principles

Strata-Log follows a few simple principles:

* Keep packages focused.
* Prefer small, clear interfaces.
* Keep dependencies minimal.
* Use context-aware operations.
* Handle errors explicitly.
* Avoid unnecessary global state.
* Document exported APIs.
* Test concurrent behavior.
* Preserve graceful shutdown behavior.
* Prefer measurable improvements over speculative optimization.

---

## Run the Application

Run directly:

```bash
go run ./cmd/strata-log
```

Or use:

```bash
make run
```

The default server listens on:

```text
http://localhost:9090
```

Health check:

```bash
curl http://localhost:9090/healthz
```

---

## Testing

Run all tests:

```bash
go test ./...
```

Or:

```bash
make test
```

Run a specific package:

```bash
go test ./internal/batcher
```

Run a specific test:

```bash
go test ./internal/batcher -run TestBatcherFlushesOnClose
```

---

## Race Detection

Concurrency is a core part of Strata-Log.

Run:

```bash
go test -race ./...
```

Or:

```bash
make race
```

Concurrent changes should pass the race detector before they are considered complete.

---

## Static Analysis

Run:

```bash
go vet ./...
```

Or:

```bash
make vet
```

---

## Formatting

Format the entire project:

```bash
gofmt -w .
```

Or:

```bash
make fmt
```

For a smaller change, format only the modified files:

```bash
gofmt -w path/to/file.go
```

---

## Makefile

The Makefile provides common development commands.

Run the complete validation pipeline:

```bash
make check
```

Available targets include:

```text
all
build
run
test
race
vet
fmt
check
benchmark
docker-build
docker-up
docker-down
clean
```

Examples:

```bash
make build
make run
make test
make race
make vet
make benchmark
```

---

## Development Workflow

A typical change should follow:

```text
Understand the requirement
          │
          ▼
Identify the affected package
          │
          ▼
Write or update tests
          │
          ▼
Implement the change
          │
          ▼
Run package tests
          │
          ▼
Run the full test suite
          │
          ▼
Run the race detector
          │
          ▼
Run go vet
          │
          ▼
Update documentation
          │
          ▼
Review the diff
          │
          ▼
Commit
```

---

## Testing Strategy

Strata-Log uses multiple levels of testing.

### Unit Tests

Unit tests live alongside the implementation they verify.

Example:

```text
internal/batcher/
├── batcher.go
└── batcher_test.go
```

### Integration Tests

Integration tests verify interactions between major components.

The current integration path covers:

```text
HTTP
 │
 ▼
Ingestion Pipeline
 │
 ▼
Batcher
 │
 ▼
SQLite
 │
 ▼
Query Service
```

Integration tests are located in:

```text
tests/
```

### Race Tests

Concurrent components should be tested with:

```bash
go test -race ./...
```

---

## Writing Tests

Tests should:

* be deterministic
* clean up resources
* avoid unnecessary sleeps
* use channels or synchronization where possible
* use context cancellation when appropriate
* cover success and failure paths
* verify concurrent behavior where relevant

Prefer synchronization primitives over arbitrary timing assumptions.

---

## Storage Development

Storage implementations are defined behind storage interfaces.

The current persistent backend is SQLite.

Storage tests are located in:

```text
internal/storage/
```

Run storage tests:

```bash
go test ./internal/storage
```

Run them with the race detector:

```bash
go test -race ./internal/storage
```

After storage changes, run:

```bash
go test ./...
```

---

## Resilience Development

The resilience package contains retry behavior used by the persistence path.

Changes to retry behavior should cover:

* immediate success
* retry after failure
* maximum attempts
* exponential backoff
* context cancellation
* invalid configuration

Run:

```bash
go test ./internal/resilience
go test -race ./internal/resilience
```

---

## Batcher Development

The batcher collects log entries and writes them in batches.

Changes should verify:

* entries are accepted
* batches flush at the configured size
* batches flush on the configured interval
* pending entries flush during shutdown
* storage failures are reported
* retry behavior works correctly
* channels close correctly
* concurrent behavior is race-free

Run:

```bash
go test ./internal/batcher
go test -race ./internal/batcher
```

---

## Pipeline Development

The pipeline processes log entries asynchronously using worker goroutines.

When changing pipeline behavior, verify:

* workers process submitted entries
* concurrent processing remains safe
* handler errors are reported
* closed pipelines reject new work
* context cancellation is respected
* shutdown does not deadlock
* worker goroutines exit correctly

Run:

```bash
go test ./internal/pipeline
go test -race ./internal/pipeline
```

---

## API Development

When changing the HTTP API:

1. Update the appropriate handler.
2. Add or update tests.
3. Update `docs/api.md`.
4. Update `README.md` when user-facing behavior changes.
5. Update `CHANGELOG.md` when appropriate.

Verify manually:

```bash
curl http://localhost:9090/healthz
```

```bash
curl http://localhost:9090/v1/logs
```

```bash
curl http://localhost:9090/metrics
```

---

## Configuration Development

Configuration changes should include:

* a sensible default
* environment variable support
* configuration tests
* documentation

Relevant files:

```text
internal/config/config.go
internal/config/config_test.go
docs/configuration.md
```

After configuration changes:

```bash
go test ./internal/config
go test ./...
```

---

## Docker Development

Build the container:

```bash
make docker-build
```

Start the deployment:

```bash
make docker-up
```

Stop the deployment:

```bash
make docker-down
```

Or use Docker Compose directly:

```bash
docker compose -f deployments/docker-compose.yml up --build
```

---

## Benchmarks

Run the project benchmarks:

```bash
make benchmark
```

Or:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

Performance changes should be measured rather than assumed.

A useful workflow is:

1. Establish a baseline.
2. Make one change.
3. Benchmark again.
4. Verify correctness.
5. Run the race detector.
6. Compare the results.

---

## Documentation

Documentation is stored in:

```text
docs/
├── api.md
├── architecture.md
├── configuration.md
├── deployment.md
└── development.md
```

Update documentation whenever user-facing behavior, configuration, API behavior, or deployment behavior changes.

---

## Security

Never commit:

* credentials
* API keys
* private certificates
* production databases
* personal data
* secret environment files

Before committing, inspect:

```bash
git status
git diff
```

Make sure no sensitive or unintended files are included.

---

## Dependency Management

Inspect dependencies:

```bash
go list -m all
```

Add or update a dependency carefully:

```bash
go get <module>
```

Then clean the module:

```bash
go mod tidy
```

Validate the project:

```bash
go test ./...
go test -race ./...
go vet ./...
```

---

## Pre-Commit Checklist

Before committing:

```bash
make check
```

Then inspect:

```bash
git status
git diff
```

Confirm:

* formatting is clean
* tests pass
* race detector passes
* `go vet` passes
* no unintended files changed
* documentation is updated when necessary

---

## Commit Messages

Use short, imperative commit messages that describe the change.

Examples:

```text
feat: add Prometheus metrics endpoint
```

```text
fix: prevent batcher shutdown deadlock
```

```text
test: add retry cancellation coverage
```

```text
docs: document HTTP API
```

```text
refactor: simplify storage query flow
```

For documentation and deployment polish:

```text
docs: polish documentation and deployment guides
```

---

## Final Validation

The standard validation sequence is:

```bash
make check
```

This runs:

```text
gofmt
go test ./...
go test -race ./...
go vet ./...
```

For container-related changes, additionally run:

```bash
make docker-build
make docker-up
```

Then verify:

```bash
curl http://localhost:9090/healthz
```

A change is ready when the relevant validation checks pass and the final Git diff contains only intentional changes.

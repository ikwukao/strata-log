# Strata-Log Development Guide

This document describes the development workflow for contributing to Strata-Log.

---

## Prerequisites

Install:

* Go
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
├── cmd/
│   └── strata-log/
├── deployments/
├── docs/
├── internal/
│   ├── batcher/
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

Strata-Log follows these principles:

* Keep packages small.
* Prefer clear interfaces.
* Keep dependencies minimal.
* Use context-aware operations.
* Avoid unnecessary global state.
* Test concurrent code with the race detector.
* Document exported APIs.
* Handle errors explicitly.
* Preserve graceful shutdown behavior.

---

## Run the Application

```bash
go run ./cmd/strata-log
```

The default server is available at:

```text
http://localhost:9090
```

---

## Run Tests

Run all tests:

```bash
go test ./...
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

Concurrency is an important part of Strata-Log.

Always run:

```bash
go test -race ./...
```

before considering a change complete.

---

## Static Analysis

Run:

```bash
go vet ./...
```

---

## Formatting

Format changed Go files:

```bash
gofmt -w path/to/file.go
```

Or format the entire project:

```bash
gofmt -w .
```

---

## Makefile

Use the project's Makefile for common workflows.

Inspect available targets:

```bash
make help
```

Use the available targets for:

* building
* testing
* race testing
* linting
* running
* container workflows

---

## Adding a New Feature

A typical feature workflow is:

```text
Understand requirement
        ↓
Identify package boundary
        ↓
Write tests
        ↓
Implement feature
        ↓
Run package tests
        ↓
Run full test suite
        ↓
Run race detector
        ↓
Run go vet
        ↓
Update documentation
        ↓
Commit
```

---

## Testing Strategy

Strata-Log uses several levels of testing.

### Unit Tests

Unit tests live alongside the package they test.

Example:

```text
internal/batcher/
├── batcher.go
└── batcher_test.go
```

### Integration Tests

Integration tests verify interactions between components such as:

```text
HTTP
 ↓
Pipeline
 ↓
Batcher
 ↓
Storage
```

### Race Tests

Concurrent components must be tested using:

```bash
go test -race ./...
```

---

## Writing Tests

Tests should:

* be deterministic
* avoid unnecessary sleeps
* use contexts with cancellation
* clean up resources
* test both success and failure paths
* verify concurrent behavior where applicable

Prefer synchronization primitives and channels over arbitrary timing assumptions.

---

## Storage Development

SQLite storage is implemented behind storage interfaces.

This allows components to depend on abstractions instead of a concrete database implementation.

When modifying storage:

```bash
go test ./internal/storage
go test -race ./internal/storage
```

Then run:

```bash
go test ./...
```

---

## Resilience Development

Changes to retry behavior should test:

* immediate success
* retry after failure
* maximum attempts
* backoff behavior
* context cancellation
* invalid configuration

Run:

```bash
go test ./internal/resilience
go test -race ./internal/resilience
```

---

## Batcher Development

The batcher is sensitive to concurrency and shutdown behavior.

Changes should verify:

* entries are accepted
* batches flush at the configured size
* batches flush on the timer
* pending entries flush during shutdown
* storage failures are reported
* retries terminate correctly
* channels close correctly
* concurrent access is race-free

Run:

```bash
go test ./internal/batcher
go test -race ./internal/batcher
```

---

## API Development

When changing the HTTP API:

1. Update the handler.
2. Add or update tests.
3. Update `docs/api.md`.
4. Update `README.md` if user-facing behavior changes.
5. Update `CHANGELOG.md`.

Verify manually:

```bash
curl http://localhost:9090/healthz
curl http://localhost:9090/v1/logs
curl http://localhost:9090/metrics
```

---

## Configuration Development

Configuration changes should include:

* a default value
* environment variable support
* configuration tests
* documentation

Update:

```text
internal/config/config.go
internal/config/config_test.go
docs/configuration.md
```

---

## Documentation

Documentation belongs in `docs/`.

Current documentation:

```text
docs/
├── api.md
├── architecture.md
├── configuration.md
├── deployment.md
└── development.md
```

Documentation should be updated alongside behavior changes.

---

## Docker Development

Build the container:

```bash
docker compose -f deployments/docker-compose.yml build
```

Start:

```bash
docker compose -f deployments/docker-compose.yml up
```

Build and start:

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Stop:

```bash
docker compose -f deployments/docker-compose.yml down
```

---

## Pre-Commit Checklist

Before committing:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

Then inspect:

```bash
git status
git diff
```

Confirm:

* no unintended files changed
* tests pass
* race detector passes
* `go vet` passes
* documentation is updated when necessary

---

## Commit Messages

Use clear, imperative commit messages.

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

---

## Pull Requests

A pull request should explain:

1. what changed
2. why it changed
3. how it was tested
4. whether documentation was updated

Example:

```markdown
## Summary

- Added configurable storage retry behavior
- Added retry metrics
- Updated deployment documentation

## Testing

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
```

---

## Debugging

Run the service:

```bash
go run ./cmd/strata-log
```

Check health:

```bash
curl http://localhost:9090/healthz
```

Check metrics:

```bash
curl http://localhost:9090/metrics
```

Check logs:

```bash
curl http://localhost:9090/v1/logs
```

For database-related issues, inspect the configured SQLite database.

---

## Performance Work

Performance changes should be measured rather than assumed.

Use:

```bash
go test -bench=. ./...
```

For focused benchmarks:

```bash
go test -bench=. ./benchmarks/...
```

When optimizing:

1. Establish a baseline.
2. Make one change.
3. Benchmark again.
4. Verify correctness.
5. Run the race detector.
6. Document meaningful improvements.

---

## Security

Do not commit:

* credentials
* API keys
* private certificates
* production databases
* personal data
* environment files containing secrets

Before submitting changes:

```bash
git status
git diff
```

Verify that sensitive files are not staged.

---

## Dependency Management

Inspect dependencies:

```bash
go list -m all
```

Update dependencies carefully:

```bash
go get <module>
go mod tidy
```

Then run:

```bash
go test ./...
go test -race ./...
go vet ./...
```

---

## Final Validation

The standard validation sequence is:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

For container changes:

```bash
docker compose -f deployments/docker-compose.yml build
docker compose -f deployments/docker-compose.yml up
```

A change is ready when the complete validation pipeline passes.

# Contributing to Strata-Log

Thank you for contributing to Strata-Log.

Strata-Log is a Go backend project focused on reliable log ingestion, asynchronous processing, persistence, resilience, and observability.

Contributions should preserve the project's simplicity, correctness, and operational reliability.

---

## Development Requirements

Install:

* Go 1.26+
* Git

Docker and Docker Compose are required when working on containerized deployment.

Verify Go:

```bash
go version
```

Verify Docker:

```bash
docker version
```

---

## Getting Started

Clone the repository:

```bash
git clone https://github.com/ikwukao/strata-log.git
cd strata-log
```

Run the service:

```bash
go run ./cmd/strata-log
```

The default server address is:

```text
http://localhost:9090
```

---

## Project Structure

```text
cmd/             Application entrypoints
internal/        Application packages
benchmarks/      Performance benchmarks
tests/           Integration tests
deployments/     Docker and deployment configuration
docs/            Project documentation
```

---

## Making Changes

Before modifying code:

1. Understand the package involved.
2. Read the existing tests.
3. Keep changes focused.
4. Avoid unnecessary dependencies.
5. Add tests for behavioral changes.
6. Update documentation when behavior changes.

---

## Code Style

Format Go code with:

```bash
gofmt -w .
```

Follow standard Go conventions.

Prefer:

* small, focused functions
* explicit error handling
* clear interfaces
* context-aware operations
* meaningful names
* documentation for exported identifiers
* simple concurrency patterns

Avoid:

* unnecessary abstractions
* global mutable state
* ignored errors without justification
* blocking operations without context handling
* introducing dependencies for functionality already available in the standard library

---

## Testing

Run the complete test suite:

```bash
go test ./...
```

Run the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

Before submitting a change, all applicable checks should pass.

---

## Integration Testing

Integration tests are located under:

```text
tests/
```

Run them with:

```bash
go test ./tests/...
```

If a change affects HTTP behavior, storage, batching, or shutdown behavior, integration coverage should be considered.

---

## Benchmarks

Performance-sensitive changes should be benchmarked.

Run:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

Do not describe performance improvements without benchmark evidence.

Benchmark results are hardware- and workload-dependent.

---

## Documentation

Update documentation when changing:

* API behavior
* configuration
* deployment
* architecture
* development workflow
* operational behavior

Documentation is located under:

```text
docs/
```

---

## Commit Messages

Use clear, imperative commit messages.

Examples:

```text
feat: add log query pagination
fix: prevent batcher shutdown deadlock
test: add storage retry coverage
docs: update deployment guide
refactor: simplify ingestion pipeline
perf: reduce log pipeline allocations
```

Keep commits focused on a single logical change whenever practical.

---

## Pull Requests

A pull request should explain:

* what changed
* why it changed
* how it was tested
* any operational implications

Include benchmark results when performance is relevant.

Before opening a pull request, run:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
git diff --check
```

---

## Security Issues

Do not publicly disclose sensitive security vulnerabilities in an issue.

Provide enough information for the vulnerability to be reproduced and understood safely.

Security-related changes should include appropriate tests whenever possible.

---

## Design Principles

Strata-Log prioritizes:

1. Correctness
2. Reliability
3. Observability
4. Simplicity
5. Performance

Performance should not be achieved by sacrificing correctness or making failure behavior unpredictable.

---

## License

By contributing to Strata-Log, you agree that your contributions will be licensed under the project's MIT License.

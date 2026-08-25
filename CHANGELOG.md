# Changelog

All notable changes to Strata-Log are documented here.

The project follows a practical Keep a Changelog-style format.

## [Unreleased]

### Added

- API documentation.
- Architecture documentation.
- Configuration documentation.
- Deployment documentation.
- Development documentation.
- Benchmark suite for ingestion and SQLite persistence.
- Prometheus-compatible `/metrics` endpoint.
- Configurable storage retry attempts and exponential backoff.

### Improved

- Graceful shutdown behavior.
- Asynchronous log ingestion pipeline.
- Batched SQLite persistence.
- Storage failure reporting.
- Runtime configuration through environment variables.
- Project documentation and development workflow.

### Reliability

- Added race-detector coverage.
- Added retry cancellation handling.
- Added atomic batch persistence.
- Added storage error reporting.
- Added graceful batcher shutdown behavior.

## [0.1.0]

### Features

- Initial Strata-Log implementation.
- HTTP log ingestion API.
- SQLite persistence.
- Log query API.
- In-memory buffering.
- Asynchronous processing pipeline.
- Configurable batching.
- Retry and resilience primitives.
- Health endpoint.
- Prometheus-compatible metrics.
- Docker deployment.
- Docker Compose configuration.
- Unit and integration tests.
- Benchmark suite.

[Unreleased]: https://github.com/ikwukao/strata-log/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ikwukao/strata-log/releases/tag/v0.1.0

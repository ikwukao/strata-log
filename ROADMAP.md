# Strata-Log Roadmap

This roadmap describes the planned evolution of Strata-Log from a lightweight
single-node logging service toward a more capable distributed logging system.

The roadmap is intentionally flexible and may change as the project evolves.

---

## Phase 1 — Core Logging Service

### Status: Complete

- [x] HTTP log ingestion
- [x] Structured log entries
- [x] Asynchronous processing
- [x] In-memory buffering
- [x] Configurable batching
- [x] SQLite persistence
- [x] Atomic batch writes
- [x] Log query API
- [x] Health endpoint
- [x] Prometheus-compatible metrics
- [x] Graceful shutdown
- [x] Retry with exponential backoff
- [x] Unit tests
- [x] Integration tests
- [x] Race-detector testing
- [x] Static analysis
- [x] Benchmarks
- [x] Docker deployment
- [x] API documentation
- [x] Architecture documentation
- [x] Configuration documentation
- [x] Deployment documentation
- [x] Development documentation

---

## Phase 2 — API Improvements

### Phase 2 status: Planned

- [ ] Pagination for log queries
- [ ] Timestamp-range filtering
- [ ] Multiple-level filtering
- [ ] Full-text message search
- [ ] Better query validation
- [ ] Consistent API error responses
- [ ] Request IDs
- [ ] API versioning improvements

---

## Phase 3 — Storage Improvements

### Phase 3 status: Planned

- [ ] Log retention policies
- [ ] Automatic database cleanup
- [ ] Storage statistics
- [ ] Configurable SQLite tuning
- [ ] Storage abstraction improvements
- [ ] Additional storage backend
- [ ] PostgreSQL support
- [ ] Object-storage archival

---

## Phase 4 — Security

### Phase 4 status: Planned

- [ ] API authentication
- [ ] API authorization
- [ ] Configurable API keys
- [ ] TLS configuration
- [ ] Request-size limits
- [ ] Rate limiting
- [ ] Security-focused integration tests
- [ ] Threat model documentation

---

## Phase 5 — Observability

### Phase 5 status: Planned

- [ ] Additional Prometheus metrics
- [ ] Batch latency metrics
- [ ] Storage latency metrics
- [ ] Query latency metrics
- [ ] Queue depth metrics
- [ ] Retry metrics
- [ ] Structured request logging
- [ ] OpenTelemetry integration
- [ ] Distributed tracing

---

## Phase 6 — Distributed Logging

### Phase 6 status: Future

- [ ] Distributed ingestion
- [ ] Horizontal scaling
- [ ] Multiple ingestion nodes
- [ ] Message broker integration
- [ ] Partitioning
- [ ] Replicated storage
- [ ] Leader/worker coordination
- [ ] Back-pressure handling across nodes
- [ ] High-availability deployment

---

## Phase 7 — Production Hardening

### Phase 7 status: Future

- [ ] Kubernetes deployment
- [ ] Helm chart
- [ ] Production configuration examples
- [ ] Automated container security scanning
- [ ] Dependency vulnerability scanning
- [ ] Load testing
- [ ] Failure-injection testing
- [ ] Capacity testing
- [ ] Disaster-recovery documentation
- [ ] Operational runbooks

---

## Guiding Principles

Future development should preserve the project's core principles:

- Keep the ingestion path fast.
- Prefer explicit concurrency over unnecessary abstraction.
- Make failures observable.
- Make shutdown predictable.
- Keep storage operations atomic.
- Keep configuration explicit.
- Test concurrent behavior.
- Measure performance instead of guessing.
- Avoid premature distributed complexity.

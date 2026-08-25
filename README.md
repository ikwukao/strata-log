# Strata-Log

A lightweight, fault-tolerant log ingestion and query service written in Go.

Strata-Log accepts structured application logs over HTTP, processes them asynchronously, batches writes, persists data to SQLite, exposes query APIs, and provides Prometheus-compatible metrics.

It is designed as a practical example of building a reliable Go service with concurrency, batching, persistence, resilience, observability, testing, and containerized deployment.

---

## Features

- Structured JSON log ingestion
- Asynchronous log processing
- Configurable in-memory buffering
- Batch persistence
- SQLite persistent storage
- Atomic batch writes
- Automatic retry with exponential backoff
- HTTP query API
- Prometheus-compatible metrics
- Health-check endpoint
- Graceful shutdown
- Configurable timeouts
- Docker and Docker Compose deployment
- Unit, integration, race, and static analysis tests
- Benchmark suite
- Zero-allocation log pipeline benchmark

---

## Architecture

```text
                    ┌──────────────────┐
                    │    HTTP Client   │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  HTTP API        │
                    │                  │
                    │ POST /v1/logs    │
                    │ GET  /v1/logs    │
                    │ GET  /healthz    │
                    │ GET  /metrics    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Ingestion        │
                    │ Pipeline         │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Buffer /         │
                    │ Workers          │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Batcher          │
                    │                  │
                    │ Size + Timer     │
                    └────────┬─────────┘
                             │
                    Retry + Exponential
                         Backoff
                             │
                             ▼
                    ┌──────────────────┐
                    │ Storage Layer    │
                    │                  │
                    │ SQLite           │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Query Service    │
                    └──────────────────┘

                    ┌──────────────────┐
                    │   Telemetry      │
                    │   /metrics       │
                    └──────────────────┘
```

See the complete architecture documentation:

- [`docs/architecture.md`](docs/architecture.md)

---

## Why Strata-Log?

Traditional application logging can become difficult to manage when applications generate large numbers of events.

Strata-Log explores a simple architecture for moving logging away from application processes while maintaining:

- fast ingestion
- controlled memory usage
- durable persistence
- failure handling
- observability
- graceful shutdown
- operational simplicity

The project is intentionally implemented in Go to demonstrate practical backend engineering concepts including concurrency, channels, worker pipelines, database transactions, HTTP services, and reliability patterns.

---

## Tech Stack

| Technology | Purpose |
| --- | --- |
| Go | Application runtime |
| SQLite | Persistent log storage |
| `modernc.org/sqlite` | Pure-Go SQLite driver |
| Docker | Containerization |
| Docker Compose | Local deployment |
| Prometheus | Metrics collection |
| `net/http` | HTTP API |
| `log/slog` | Structured application logging |

---

## Requirements

For local development:

- Go 1.26+
- Git

For containerized deployment:

- Docker
- Docker Compose

---

## Quick Start

Clone the repository:

```bash
git clone https://github.com/ikwukao/strata-log.git
cd strata-log
```

Run the service:

```bash
go run ./cmd/strata-log
```

By default, Strata-Log listens on:

```text
http://localhost:9090
```

Health check:

```bash
curl http://localhost:9090/healthz
```

Expected response:

```json
{"status":"ok"}
```

---

## Ingest a Log

Strata-Log accepts structured JSON log entries through:

```text
POST /v1/logs
```

Example:

```bash
curl -X POST http://localhost:9090/v1/logs \
  -H 'Content-Type: application/json' \
  -d '{
    "timestamp": "2026-08-25T19:00:00Z",
    "level": "info",
    "service": "api",
    "message": "Strata-Log is alive",
    "fields": {
      "environment": "development"
    }
  }'
```

The `timestamp` field is required.

---

## Query Logs

Retrieve persisted logs:

```bash
curl http://localhost:9090/v1/logs
```

Filter by level:

```bash
curl 'http://localhost:9090/v1/logs?level=error'
```

Filter by service:

```bash
curl 'http://localhost:9090/v1/logs?service=api'
```

Limit the number of results:

```bash
curl 'http://localhost:9090/v1/logs?limit=20'
```

Combine filters:

```bash
curl 'http://localhost:9090/v1/logs?level=error&service=api&limit=50'
```

Example response:

```json
{
  "logs": [
    {
      "id": 1,
      "timestamp": "2026-08-25T19:00:00Z",
      "level": "info",
      "service": "api",
      "message": "Strata-Log is alive",
      "fields": {
        "environment": "development"
      }
    }
  ],
  "count": 1
}
```

See the complete API reference:

- [`docs/api.md`](docs/api.md)

---

## Metrics

Strata-Log exposes Prometheus-compatible metrics at:

```text
GET /metrics
```

Example:

```bash
curl http://localhost:9090/metrics
```

Available counters include:

```text
strata_log_ingested_total
strata_log_stored_total
strata_log_errors_total
```

These counters provide basic visibility into ingestion, persistence, and processing failures.

---

## Configuration

Strata-Log is configured through environment variables.

Important configuration values include:

```text
STRATA_LOG_HOST
STRATA_LOG_PORT
STRATA_LOG_STORAGE_PATH
STRATA_LOG_BUFFER_CAPACITY
STRATA_LOG_BATCH_SIZE
STRATA_LOG_FLUSH_PERIOD
STRATA_LOG_SHUTDOWN_TIMEOUT
STRATA_LOG_RETRY_ATTEMPTS
STRATA_LOG_RETRY_BACKOFF
```

Default server configuration:

```text
Host:            0.0.0.0
Port:            9090
Read timeout:    10s
Write timeout:   10s
Idle timeout:    60s
Shutdown timeout: 10s
```

See:

- [`docs/configuration.md`](docs/configuration.md)

---

## Docker

Build and start Strata-Log with Docker Compose:

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Run in the background:

```bash
docker compose -f deployments/docker-compose.yml up --build -d
```

Check the service:

```bash
curl http://localhost:9090/healthz
```

Stop the deployment:

```bash
docker compose -f deployments/docker-compose.yml down
```

Deployment documentation:

- [`docs/deployment.md`](docs/deployment.md)

---

## Project Structure

```text
strata-log/
├── benchmarks/
│   ├── ingest_test.go
│   └── storage_test.go
│
├── cmd/
│   └── strata-log/
│       └── main.go
│
├── deployments/
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── prometheus.yml
│
├── docs/
│   ├── api.md
│   ├── architecture.md
│   ├── configuration.md
│   ├── deployment.md
│   └── development.md
│
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
│
├── tests/
│   └── integration_test.go
│
├── CONTRIBUTING.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── ROADMAP.md
└── README.md
```

---

## Reliability

Strata-Log is designed to avoid losing an entire batch because of a transient storage failure.

The persistence path uses:

1. Batch accumulation
2. Atomic database transactions
3. Configurable retry attempts
4. Exponential retry backoff
5. Context-aware cancellation
6. Error reporting through the batcher
7. Graceful shutdown

The retry mechanism stops immediately when its context is canceled.

---

## Graceful Shutdown

Strata-Log handles:

```text
SIGINT
SIGTERM
```

During shutdown it:

1. Stops accepting new HTTP requests.
2. Shuts down the HTTP server.
3. Stops the ingestion processor.
4. Flushes pending batches.
5. Closes the storage backend.
6. Exits cleanly.

---

## Testing

Run the complete test suite:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Run formatting:

```bash
gofmt -w .
```

Run benchmarks:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

---

## Benchmarks

Benchmarks were run locally on:

```text
OS:   Linux
Arch: amd64
CPU:  Intel(R) Core(TM) i5-3427U CPU @ 1.80GHz
```

Repeated benchmark runs produced the following ranges:

| Benchmark | Result |
| --- | ---: |
| `BenchmarkLogPipeline` | ~1.03–1.22 µs/op |
| `BenchmarkLogPipeline` allocations | 0 B/op |
| `BenchmarkLogPipeline` allocations | 0 allocs/op |
| `BenchmarkSQLiteWriteBatch` | ~10.1–11.9 ms/op |
| `BenchmarkSQLiteWriteBatch` memory | ~63 KB/op |
| `BenchmarkSQLiteWriteBatch` allocations | ~2,012 allocs/op |

Example:

```text
BenchmarkLogPipeline-2          1000000    1139 ns/op       0 B/op       0 allocs/op
BenchmarkSQLiteWriteBatch-2         122    9379326 ns/op   63070 B/op   2012 allocs/op
```

These results are **local reference measurements**, not production performance guarantees. Actual performance will depend on hardware, workload, SQLite configuration, batch size, filesystem, and deployment environment.

---

## Development

Development documentation covers:

- local setup
- project structure
- development workflow
- testing
- formatting
- benchmarks

See:

- [`docs/development.md`](docs/development.md)

---

## API Documentation

Complete API documentation is available at:

- [`docs/api.md`](docs/api.md)

Primary endpoints:

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Health check |
| `POST` | `/v1/logs` | Ingest log |
| `GET` | `/v1/logs` | Query logs |
| `GET` | `/metrics` | Prometheus metrics |

---

## Roadmap

The current roadmap focuses on evolving Strata-Log toward a more capable distributed logging platform.

Potential future improvements include:

- richer query filters
- pagination
- log retention policies
- additional storage backends
- improved metrics
- authentication and authorization
- distributed ingestion
- horizontal scaling
- improved observability
- production deployment examples

See [`ROADMAP.md`](ROADMAP.md) for the current roadmap.

---

## Contributing

Contributions are welcome.

Before submitting changes:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

Please read:

- [`CONTRIBUTING.md`](CONTRIBUTING.md)

---

## License

Strata-Log is licensed under the MIT License.

See [`LICENSE`](LICENSE) for the full license text.

---

## Project Status

Strata-Log is an actively developed Go backend project demonstrating:

- concurrent processing
- asynchronous pipelines
- batching
- persistent storage
- retry and resilience patterns
- HTTP API design
- observability
- automated testing
- containerized deployment

The current implementation is suitable for local development, experimentation, and demonstration of backend engineering practices.

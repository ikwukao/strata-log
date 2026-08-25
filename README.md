# Strata-Log

A lightweight, fault-tolerant log ingestion and query service written in Go.

Strata-Log accepts structured application logs over HTTP, processes them asynchronously, batches writes, persists them to SQLite, exposes query endpoints, and provides Prometheus-compatible metrics.

The project is designed as a practical Go backend system demonstrating concurrency, asynchronous processing, batching, persistence, retry handling, observability, graceful shutdown, testing, and containerized deployment.

---

## Features

* Structured JSON log ingestion
* Asynchronous log processing
* Concurrent worker pipeline
* Bounded in-memory buffering
* Configurable batch processing
* SQLite persistent storage
* Atomic batch writes
* Retry with exponential backoff
* HTTP log query API
* Prometheus-compatible metrics
* Health-check endpoint
* Graceful shutdown
* Configurable runtime settings
* Docker and Docker Compose support
* Unit and integration tests
* Race-detector coverage
* Static analysis with `go vet`
* Benchmark suite

---

## Architecture

```text
                         HTTP Client
                              │
                              ▼
                    ┌──────────────────┐
                    │    HTTP API      │
                    │                  │
                    │ POST /v1/logs    │
                    │ GET  /v1/logs    │
                    │ GET  /healthz    │
                    │ GET  /metrics    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │    Ingestion     │
                    │    Pipeline      │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Workers / Buffer │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │     Batcher      │
                    │                  │
                    │ Size + Interval  │
                    └────────┬─────────┘
                             │
                       Retry + Backoff
                             │
                             ▼
                    ┌──────────────────┐
                    │     SQLite       │
                    │    Storage       │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Query Service   │
                    └──────────────────┘

                    ┌──────────────────┐
                    │    Telemetry     │
                    │    /metrics      │
                    └──────────────────┘
```

For a deeper explanation of the system design, see [`docs/architecture.md`](docs/architecture.md).

---

## Tech Stack

| Technology           | Purpose                        |
| -------------------- | ------------------------------ |
| Go                   | Application runtime            |
| `net/http`           | HTTP server and API            |
| SQLite               | Persistent log storage         |
| `modernc.org/sqlite` | Pure-Go SQLite driver          |
| Docker               | Containerization               |
| Docker Compose       | Local container deployment     |
| Prometheus           | Metrics collection             |
| `log/slog`           | Structured application logging |

---

## Requirements

### Local Development

* Go 1.26+
* Git

### Containerized Development

* Docker
* Docker Compose

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

Strata-Log listens on port `9090` by default.

### Health Check

```bash
curl http://localhost:9090/healthz
```

Expected response:

```json
{
  "status": "ok"
}
```

---

## Ingest a Log

Send a structured log entry:

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

Limit results:

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

See the complete API reference in [`docs/api.md`](docs/api.md).

---

## Metrics

Prometheus-compatible metrics are exposed at:

```text
GET /metrics
```

View the metrics:

```bash
curl http://localhost:9090/metrics
```

The service exposes counters including:

```text
strata_log_ingested_total
strata_log_stored_total
strata_log_errors_total
```

---

## Configuration

Strata-Log is configured through environment variables.

Common settings include:

```text
STRATA_LOG_HOST
STRATA_LOG_PORT
STRATA_LOG_STORAGE_PATH
STRATA_LOG_RETRY_ATTEMPTS
STRATA_LOG_RETRY_BACKOFF
STRATA_LOG_BUFFER_CAPACITY
STRATA_LOG_BATCH_SIZE
STRATA_LOG_FLUSH_PERIOD
STRATA_LOG_READ_TIMEOUT
STRATA_LOG_WRITE_TIMEOUT
STRATA_LOG_IDLE_TIMEOUT
STRATA_LOG_SHUTDOWN_TIMEOUT
```

Default server configuration:

```text
Host:             0.0.0.0
Port:             9090
Read timeout:     10s
Write timeout:    10s
Idle timeout:     60s
Shutdown timeout: 10s
```

See [`docs/configuration.md`](docs/configuration.md) for the complete configuration reference.

---

## Docker

Build the image:

```bash
make docker-build
```

Start Strata-Log:

```bash
make docker-up
```

Or use Docker Compose directly:

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Run in the background:

```bash
docker compose -f deployments/docker-compose.yml up --build -d
```

Stop the deployment:

```bash
make docker-down
```

Verify the service:

```bash
curl http://localhost:9090/healthz
```

See [`docs/deployment.md`](docs/deployment.md) for deployment instructions.

---

## Makefile

Strata-Log includes a Makefile for common development, testing, build, Docker, and operational workflows.

Display all available commands:

```bash
make help
```

### Development

Format Go source files:

```bash
make fmt
```

Run tests:

```bash
make test
```

Run tests with the race detector:

```bash
make test-race
```

Run static analysis:

```bash
make vet
```

Build the Strata-Log binary:

```bash
make build
```

Run the complete validation and build workflow:

```bash
make check
```

### Docker Compose

Build the Docker image:

```bash
make docker-build
```

Start the Docker Compose stack:

```bash
make up
```

Stop the stack:

```bash
make down
```

Restart the stack:

```bash
make restart
```

Follow Strata-Log container logs:

```bash
make logs
```

### Operations

Check service health:

```bash
make health
```

Display Prometheus metrics:

```bash
make metrics
```

### Cleanup

Remove local build artifacts:

```bash
make clean
```

The Makefile provides a consistent interface for local development and container-based workflows, while the underlying commands remain directly available when needed.

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
├── README.md
├── go.mod
└── go.sum
```

---

## Reliability

The persistence path is designed around several reliability mechanisms:

1. Bounded ingestion buffering
2. Concurrent processing
3. Batch accumulation
4. Atomic SQLite batch writes
5. Configurable retry attempts
6. Exponential retry backoff
7. Context-aware cancellation
8. Graceful shutdown

Transient storage failures can therefore be retried before being reported through the batcher's error channel.

---

## Graceful Shutdown

Strata-Log handles:

```text
SIGINT
SIGTERM
```

During shutdown, the service:

1. Stops the HTTP server.
2. Stops accepting new processing work.
3. Drains pending pipeline work.
4. Flushes pending batches.
5. Closes the storage backend.
6. Exits cleanly.

The shutdown timeout is configurable through:

```text
STRATA_LOG_SHUTDOWN_TIMEOUT
```

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

Run benchmarks:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

Or use:

```bash
make check
```

---

## Development Docs

Development documentation covers:

* local setup
* project structure
* testing
* concurrency
* storage development
* resilience development
* batcher development
* API development
* Docker workflows

See [`docs/development.md`](docs/development.md).

---

## API

| Method | Endpoint   | Purpose            |
| ------ | ---------- | ------------------ |
| `GET`  | `/healthz` | Health check       |
| `POST` | `/v1/logs` | Ingest a log       |
| `GET`  | `/v1/logs` | Query logs         |
| `GET`  | `/metrics` | Prometheus metrics |

See [`docs/api.md`](docs/api.md) for the complete API reference.

---

## Project Status

Strata-Log is an MVP demonstrating practical Go backend engineering patterns, including:

* concurrent processing
* asynchronous ingestion
* bounded buffering
* batching
* persistent storage
* retry and resilience patterns
* HTTP API design
* observability
* graceful shutdown
* automated testing
* containerized deployment

The project is intended for local development, experimentation, learning, and demonstration of backend engineering practices.

---

## Contributing

Contributions are welcome.

Before submitting changes, run:

```bash
make check
```

Please read [`CONTRIBUTING.md`](CONTRIBUTING.md) before contributing.

---

## License

Strata-Log is licensed under the MIT License.

See [`LICENSE`](LICENSE) for the full license text.

# Strata-Log Architecture

Strata-Log is a lightweight, fault-tolerant log ingestion and storage service written in Go.

Its architecture separates HTTP ingestion, asynchronous processing, batching, persistence, querying, resilience, and observability.

---

## High-Level Architecture

```text
                         ┌─────────────────────┐
                         │       Clients       │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │     HTTP Server     │
                         │                     │
                         │  /healthz           │
                         │  /v1/logs           │
                         │  /metrics           │
                         └──────────┬──────────┘
                                    │
                         POST /v1/logs
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │   Ingestion Layer   │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │      Pipeline       │
                         │  Worker Processing  │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │      Batcher        │
                         │  Size + Time Flush  │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │     Resilience      │
                         │    Retry + Backoff  │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │   SQLite Storage    │
                         └─────────────────────┘

                         GET /v1/logs
                              │
                              ▼
                         ┌──────────────┐
                         │ Query Layer  │
                         └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │    SQLite    │
                         └──────────────┘
```

---

## Core Components

### HTTP Server

The HTTP server is responsible for exposing Strata-Log's public API.

Endpoints include:

* `GET /healthz`
* `POST /v1/logs`
* `GET /v1/logs`
* `GET /metrics`

The server uses Go's standard `net/http` package.

---

## Ingestion

The ingestion layer validates incoming log requests and converts them into internal `LogEntry` values.

The handler does not directly write to SQLite.

Instead, entries are submitted to the asynchronous processing pipeline.

This keeps request latency independent from database write latency.

---

## Pipeline

The pipeline provides asynchronous processing using worker goroutines.

Its responsibilities include:

* receiving accepted log entries
* distributing work across workers
* forwarding entries to the batcher
* propagating processing failures

This provides controlled concurrency and prevents unbounded goroutine creation.

---

## Batcher

The batcher groups individual log entries into batches.

A batch is flushed when either:

1. the configured batch size is reached, or
2. the flush interval expires.

For example:

```text
Batch Size  = 100
Flush Period = 500ms
```

A batch containing 100 entries is immediately flushed.

If fewer than 100 entries arrive, the batch is flushed after 500ms.

Batching reduces database transaction overhead and improves write throughput.

---

## Resilience

Storage writes are protected by retry logic.

The retry mechanism:

1. performs the first attempt immediately
2. retries failed operations
3. applies exponential backoff
4. stops after the configured retry count
5. respects context cancellation

Example:

```text
Attempt 1
   │
   ├── failure
   ▼
Backoff
   │
   ▼
Attempt 2
   │
   ├── failure
   ▼
Longer Backoff
   │
   ▼
Attempt 3
```

This protects the system against temporary storage failures without retrying indefinitely.

---

## Storage

Strata-Log currently uses SQLite as its persistent storage backend.

The storage layer provides:

* atomic batch writes
* structured field serialization
* indexed queries
* context-aware operations
* WAL mode for improved concurrent access

The database contains the `logs` table.

Indexes exist for:

* timestamp
* level
* service

---

## Query Layer

The query layer provides read access to persisted logs.

Supported filters include:

* log level
* service
* result limit

Queries are executed against SQLite and returned as JSON.

The query layer is intentionally separated from ingestion so reads and writes remain independently structured.

---

## Telemetry

Strata-Log exposes application counters through `/metrics`.

Current metrics include:

```text
strata_log_ingested_total
strata_log_stored_total
strata_log_errors_total
```

Counters use atomic operations and are safe for concurrent updates.

The endpoint follows the Prometheus text exposition format.

---

## Graceful Shutdown

Strata-Log handles:

* `SIGINT`
* `SIGTERM`

Shutdown follows this sequence:

```text
Signal
  ↓
Stop HTTP server
  ↓
Stop accepting new logs
  ↓
Close processor
  ↓
Flush batcher
  ↓
Close SQLite
  ↓
Exit
```

A shutdown timeout prevents the process from waiting indefinitely.

---

## Concurrency Model

The system uses Go goroutines and channels for asynchronous processing.

Major concurrent components include:

* HTTP handlers
* pipeline workers
* batcher worker
* telemetry counters
* storage operations

Shared counters use atomic operations.

The project is tested with the Go race detector.

---

## Data Flow

### Write Path

```text
Client
  ↓
POST /v1/logs
  ↓
Validation
  ↓
Pipeline
  ↓
Worker
  ↓
Batcher
  ↓
Retry
  ↓
SQLite transaction
```

### Read Path

```text
Client
  ↓
GET /v1/logs
  ↓
Query Handler
  ↓
Query Service
  ↓
SQLite
  ↓
JSON response
```

---

## Package Structure

```text
internal/
├── batcher/
├── config/
├── ingest/
├── pipeline/
├── query/
├── resilience/
├── storage/
└── telemetry/
```

### `batcher`

Groups entries and persists batches.

### `config`

Loads and validates runtime configuration.

### `ingest`

Handles HTTP log ingestion and health checks.

### `pipeline`

Provides asynchronous log processing.

### `query`

Provides HTTP querying of persisted logs.

### `resilience`

Contains reliability primitives such as retries.

### `storage`

Provides persistence abstractions and SQLite implementation.

### `telemetry`

Provides application metrics.

---

## Design Principles

Strata-Log follows several core principles:

* asynchronous ingestion
* bounded buffering
* batch persistence
* explicit failure handling
* context-aware operations
* graceful shutdown
* observable runtime behavior
* standard Go libraries where practical
* small, independently testable packages

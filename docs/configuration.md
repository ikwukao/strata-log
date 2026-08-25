# Strata-Log Configuration

Strata-Log is configured through environment variables.

All configuration values have safe defaults, allowing the service to run without a configuration file.

---

## Configuration Variables

### Server

| Variable                      | Default   | Description                     |
| ----------------------------- | --------- | ------------------------------- |
| `STRATA_LOG_HOST`             | `0.0.0.0` | HTTP server host                |
| `STRATA_LOG_PORT`             | `9090`    | HTTP server port                |
| `STRATA_LOG_READ_TIMEOUT`     | `10s`     | Maximum request read duration   |
| `STRATA_LOG_WRITE_TIMEOUT`    | `10s`     | Maximum response write duration |
| `STRATA_LOG_IDLE_TIMEOUT`     | `60s`     | Keep-alive idle timeout         |
| `STRATA_LOG_SHUTDOWN_TIMEOUT` | `10s`     | Graceful shutdown timeout       |

---

## Storage

| Variable                    | Default         | Description                    |
| --------------------------- | --------------- | ------------------------------ |
| `STRATA_LOG_STORAGE_PATH`   | `strata-log.db` | SQLite database path           |
| `STRATA_LOG_RETRY_ATTEMPTS` | `3`             | Maximum storage retry attempts |
| `STRATA_LOG_RETRY_BACKOFF`  | `100ms`         | Initial retry backoff          |

---

## Buffering

| Variable                     | Default | Description                  |
| ---------------------------- | ------: | ---------------------------- |
| `STRATA_LOG_BUFFER_CAPACITY` | `10000` | Maximum buffered log entries |

The buffer provides bounded asynchronous ingestion.

A larger buffer can absorb traffic bursts but consumes more memory.

---

## Batching

| Variable                  | Default | Description                          |
| ------------------------- | ------: | ------------------------------------ |
| `STRATA_LOG_BATCH_SIZE`   |   `100` | Maximum entries per batch            |
| `STRATA_LOG_FLUSH_PERIOD` | `500ms` | Maximum time before flushing a batch |

A batch is flushed when either the maximum size or flush period is reached.

---

## Example Configuration

```bash
export STRATA_LOG_HOST=0.0.0.0
export STRATA_LOG_PORT=9090

export STRATA_LOG_STORAGE_PATH=/var/lib/strata-log/strata-log.db

export STRATA_LOG_BUFFER_CAPACITY=10000

export STRATA_LOG_BATCH_SIZE=100
export STRATA_LOG_FLUSH_PERIOD=500ms

export STRATA_LOG_RETRY_ATTEMPTS=3
export STRATA_LOG_RETRY_BACKOFF=100ms

export STRATA_LOG_READ_TIMEOUT=10s
export STRATA_LOG_WRITE_TIMEOUT=10s
export STRATA_LOG_IDLE_TIMEOUT=60s
export STRATA_LOG_SHUTDOWN_TIMEOUT=10s
```

Then start Strata-Log:

```bash
go run ./cmd/strata-log
```

---

## Docker Configuration

Environment variables can be supplied through Docker Compose.

Example:

```yaml
services:
  strata-log:
    environment:
      STRATA_LOG_PORT: "9090"
      STRATA_LOG_STORAGE_PATH: "/data/strata-log.db"
      STRATA_LOG_BATCH_SIZE: "100"
      STRATA_LOG_FLUSH_PERIOD: "500ms"
      STRATA_LOG_RETRY_ATTEMPTS: "3"
      STRATA_LOG_RETRY_BACKOFF: "100ms"
```

---

## Configuration Validation

Configuration values are parsed from environment variables.

Invalid integer or duration values fall back to their configured defaults.

For example:

```text
STRATA_LOG_PORT=invalid
```

falls back to:

```text
9090
```

Similarly:

```text
STRATA_LOG_FLUSH_PERIOD=invalid
```

falls back to:

```text
500ms
```

---

## Duration Format

Duration values use Go's duration format.

Examples:

```text
500ms
1s
5s
1m
5m
```

---

## Recommended Production Settings

A starting point for a small production deployment:

```bash
STRATA_LOG_HOST=0.0.0.0
STRATA_LOG_PORT=9090

STRATA_LOG_BUFFER_CAPACITY=10000

STRATA_LOG_BATCH_SIZE=100
STRATA_LOG_FLUSH_PERIOD=500ms

STRATA_LOG_RETRY_ATTEMPTS=3
STRATA_LOG_RETRY_BACKOFF=100ms

STRATA_LOG_READ_TIMEOUT=10s
STRATA_LOG_WRITE_TIMEOUT=10s
STRATA_LOG_IDLE_TIMEOUT=60s
STRATA_LOG_SHUTDOWN_TIMEOUT=10s
```

Tune these values according to workload, available memory, storage performance, and expected traffic.

---

## Environment Inspection

To inspect the current configuration environment:

```bash
env | grep '^STRATA_LOG_'
```

---

## Important Notes

### Storage Path

The process must have permission to create and write to the configured SQLite database path.

For production deployments, prefer a persistent filesystem location.

### Buffer Capacity

Increasing the buffer increases burst tolerance but also increases memory usage.

### Batch Size

Larger batches generally reduce transaction overhead but may increase latency.

### Flush Period

A shorter flush period reduces persistence latency but can increase database activity.

### Retry Attempts

Higher retry counts improve tolerance to temporary failures but can delay failure reporting.

### Retry Backoff

The configured backoff is the initial delay. Subsequent retry delays increase exponentially.

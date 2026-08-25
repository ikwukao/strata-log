# Strata-Log Configuration

Strata-Log is configured through environment variables.

If an environment variable is not set or contains an invalid value, the application uses its configured default.

---

## Server Configuration

| Variable                      | Default   | Description                                  |
| ----------------------------- | --------- | -------------------------------------------- |
| `STRATA_LOG_HOST`             | `0.0.0.0` | HTTP server bind address                     |
| `STRATA_LOG_PORT`             | `9090`    | HTTP server port                             |
| `STRATA_LOG_READ_TIMEOUT`     | `10s`     | Maximum HTTP request read duration           |
| `STRATA_LOG_WRITE_TIMEOUT`    | `10s`     | Maximum HTTP response write duration         |
| `STRATA_LOG_IDLE_TIMEOUT`     | `60s`     | Maximum time to keep an idle connection open |
| `STRATA_LOG_SHUTDOWN_TIMEOUT` | `10s`     | Maximum graceful shutdown duration           |

Example:

```bash
export STRATA_LOG_HOST=127.0.0.1
export STRATA_LOG_PORT=9090
export STRATA_LOG_READ_TIMEOUT=10s
export STRATA_LOG_WRITE_TIMEOUT=10s
export STRATA_LOG_IDLE_TIMEOUT=60s
export STRATA_LOG_SHUTDOWN_TIMEOUT=10s
```

---

## Storage Configuration

| Variable                            | Default         | Description                    |
| ----------------------------------- | --------------- | ------------------------------ |
| `STRATA_LOG_STORAGE_PATH`           | `strata-log.db` | SQLite database path           |
| `STRATA_LOG_STORAGE_RETRY_ATTEMPTS` | `3`             | Maximum storage write attempts |
| `STRATA_LOG_STORAGE_RETRY_BACKOFF`  | `100ms`         | Initial retry backoff duration |

Retry backoff is exponential.

For example, with:

```text
STRATA_LOG_STORAGE_RETRY_BACKOFF=100ms
```

successive retry delays are approximately:

```text
100ms
200ms
400ms
...
```

Example:

```bash
export STRATA_LOG_STORAGE_PATH=/data/strata-log.db
export STRATA_LOG_STORAGE_RETRY_ATTEMPTS=3
export STRATA_LOG_STORAGE_RETRY_BACKOFF=100ms
```

---

## Buffer Configuration

| Variable                     | Default | Description                                                    |
| ---------------------------- | ------: | -------------------------------------------------------------- |
| `STRATA_LOG_BUFFER_CAPACITY` | `10000` | Maximum number of entries buffered for asynchronous processing |

Example:

```bash
export STRATA_LOG_BUFFER_CAPACITY=10000
```

---

## Batcher Configuration

| Variable                  | Default | Description                                    |
| ------------------------- | ------: | ---------------------------------------------- |
| `STRATA_LOG_BATCH_SIZE`   |   `100` | Maximum number of entries written in one batch |
| `STRATA_LOG_FLUSH_PERIOD` | `500ms` | Maximum interval between batch flushes         |

A batch is flushed when either:

* the configured batch size is reached
* the flush interval expires
* the application shuts down

Example:

```bash
export STRATA_LOG_BATCH_SIZE=100
export STRATA_LOG_FLUSH_PERIOD=500ms
```

---

## Example Configuration

A complete local configuration might look like:

```bash
export STRATA_LOG_HOST=0.0.0.0
export STRATA_LOG_PORT=9090

export STRATA_LOG_STORAGE_PATH=./data/strata-log.db
export STRATA_LOG_STORAGE_RETRY_ATTEMPTS=3
export STRATA_LOG_STORAGE_RETRY_BACKOFF=100ms

export STRATA_LOG_BUFFER_CAPACITY=10000

export STRATA_LOG_BATCH_SIZE=100
export STRATA_LOG_FLUSH_PERIOD=500ms

export STRATA_LOG_READ_TIMEOUT=10s
export STRATA_LOG_WRITE_TIMEOUT=10s
export STRATA_LOG_IDLE_TIMEOUT=60s
export STRATA_LOG_SHUTDOWN_TIMEOUT=10s
```

Start the application:

```bash
go run ./cmd/strata-log
```

---

## Docker Configuration

When running inside Docker, configure the database path to point to the mounted persistent storage location.

Example:

```text
STRATA_LOG_STORAGE_PATH=/data/strata-log.db
```

The corresponding Docker volume can be mounted at:

```text
/data
```

---

## Invalid Configuration

Invalid integer and duration values fall back to their configured defaults.

For example:

```bash
export STRATA_LOG_PORT=invalid
```

causes the application to use:

```text
9090
```

Similarly:

```bash
export STRATA_LOG_FLUSH_PERIOD=invalid
```

falls back to:

```text
500ms
```

Environment variables should therefore be validated before deployment.

---

## Configuration Source

Configuration is implemented in:

```text
internal/config/config.go
```

Configuration behavior is tested in:

```text
internal/config/config_test.go
```

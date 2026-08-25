# Strata-Log API

Strata-Log exposes a small HTTP API for health checks, log ingestion, log querying, and Prometheus-compatible metrics.

## Base URL

By default:

```text
http://localhost:9090
```

The server address can be changed through environment variables. See [configuration.md](configuration.md).

---

## Endpoints

| Method | Endpoint   | Description                   |
| ------ | ---------- | ----------------------------- |
| `GET`  | `/healthz` | Health check                  |
| `POST` | `/v1/logs` | Ingest a log entry            |
| `GET`  | `/v1/logs` | Query persisted logs          |
| `GET`  | `/metrics` | Prometheus-compatible metrics |

---

## Health Check

### Health check request

```bash
curl http://localhost:9090/healthz
```

### Health check response

```json
{
  "status": "ok"
}
```

A successful health check returns HTTP `200`.

---

## Log Ingestion

### `POST /v1/logs`

Accepts a single log entry and submits it to the asynchronous processing pipeline.

### Ingestion request

```bash
curl -X POST http://localhost:9090/v1/logs \
  -H 'Content-Type: application/json' \
  -d '{
    "timestamp": "2026-08-25T12:00:00Z",
    "level": "info",
    "service": "api",
    "message": "Strata-Log is alive",
    "fields": {
      "environment": "development"
    }
  }'
```

### Request fields

| Field       | Type   | Required | Description                    |
| ----------- | ------ | -------- | ------------------------------ |
| `timestamp` | string | Yes      | RFC3339 timestamp              |
| `level`     | string | Yes      | Log severity                   |
| `service`   | string | Yes      | Source service                 |
| `message`   | string | Yes      | Log message                    |
| `fields`    | object | No       | Additional structured metadata |

### Ingestion response

```json
{
  "status": "accepted"
}
```

The ingestion endpoint is asynchronous. A successful response means the log was accepted by the processing pipeline, not necessarily that it has already been persisted to SQLite.

---

## Query Logs

### `GET /v1/logs`

Returns persisted log entries.

### Query parameters

| Parameter | Type    | Description               |
| --------- | ------- | ------------------------- |
| `level`   | string  | Filter by log level       |
| `service` | string  | Filter by service         |
| `limit`   | integer | Maximum number of records |

The default limit is `100`.

The maximum limit is `1000`.

### Get recent logs

```bash
curl http://localhost:9090/v1/logs
```

### Filter by level

```bash
curl 'http://localhost:9090/v1/logs?level=error'
```

### Filter by service

```bash
curl 'http://localhost:9090/v1/logs?service=api'
```

### Combine filters

```bash
curl 'http://localhost:9090/v1/logs?level=error&service=api&limit=50'
```

### Query response

```json
{
  "logs": [
    {
      "id": 1,
      "timestamp": "2026-08-25T12:00:00Z",
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

Results are ordered newest first.

---

## Metrics

### `GET /metrics`

Exposes application metrics using the Prometheus text exposition format.

```bash
curl http://localhost:9090/metrics
```

### Metrics example

```text
# HELP strata_log_ingested_total Total accepted log entries.
# TYPE strata_log_ingested_total counter
strata_log_ingested_total 10

# HELP strata_log_stored_total Total persisted log entries.
# TYPE strata_log_stored_total counter
strata_log_stored_total 10

# HELP strata_log_errors_total Total processing and storage errors.
# TYPE strata_log_errors_total counter
strata_log_errors_total 0
```

### Available metrics

* `strata_log_ingested_total`
* `strata_log_stored_total`
* `strata_log_errors_total`

---

## HTTP Status Codes

| Status | Meaning                                  |
| ------ | ---------------------------------------- |
| `200`  | Request succeeded                        |
| `202`  | Log accepted for asynchronous processing |
| `400`  | Invalid request                          |
| `405`  | HTTP method not allowed                  |
| `500`  | Internal server error                    |
| `503`  | Service temporarily unavailable          |

---

## Error Responses

Errors are returned as plain-text HTTP error responses.

### Error response example

```text
timestamp is required
```

Clients should treat non-2xx responses as failed requests.

---

## Design Notes

The ingestion path is intentionally asynchronous:

```text
HTTP Request
    ↓
Ingestion Handler
    ↓
Pipeline
    ↓
Batcher
    ↓
Retry
    ↓
SQLite
```

Queries execute directly against persisted SQLite data.

This separation allows ingestion to remain fast while storage is handled asynchronously.

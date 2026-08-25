# Strata-Log Deployment Guide

This document describes how to build, run, verify, and operate Strata-Log using Docker.

Strata-Log uses SQLite for persistent storage, so deployments should provide persistent storage for the database file.

---

## Prerequisites

Install:

* Docker
* Docker Compose

Verify Docker:

```bash
docker version
```

Verify Docker Compose:

```bash
docker compose version
```

---

## Build the Image

Using the Makefile:

```bash
make docker-build
```

Or directly:

```bash
docker build \
  -f deployments/Dockerfile \
  -t strata-log:latest \
  .
```

---

## Start with Docker Compose

Build and start:

```bash
docker compose \
  -f deployments/docker-compose.yml \
  up --build
```

Run in the background:

```bash
docker compose \
  -f deployments/docker-compose.yml \
  up --build -d
```

Check the running services:

```bash
docker compose \
  -f deployments/docker-compose.yml \
  ps
```

---

## Stop the Deployment

Stop the containers:

```bash
docker compose \
  -f deployments/docker-compose.yml \
  down
```

The exact persistence behavior depends on the volumes configured in the Compose file.

---

## Health Check

Once Strata-Log is running:

```bash
curl http://localhost:9090/healthz
```

Expected response:

```json
{"status":"ok"}
```

---

## Ingest a Log

Verify ingestion:

```bash
curl -X POST http://localhost:9090/v1/logs \
  -H 'Content-Type: application/json' \
  -d '{
    "timestamp": "2026-08-25T12:00:00Z",
    "level": "info",
    "service": "deployment",
    "message": "deployment verification"
  }'
```

---

## Query Logs

Verify that the log can be queried:

```bash
curl http://localhost:9090/v1/logs
```

---

## Metrics

Strata-Log exposes Prometheus-compatible metrics:

```bash
curl http://localhost:9090/metrics
```

The deployment includes:

```text
deployments/prometheus.yml
```

A Prometheus configuration can scrape the service using:

```yaml
scrape_configs:
  - job_name: strata-log
    static_configs:
      - targets:
          - strata-log:9090
```

---

## Persistent SQLite Storage

SQLite stores the database in the path configured by:

```text
STRATA_LOG_STORAGE_PATH
```

For a container deployment, the database should be stored on a mounted persistent volume.

Example:

```text
/data/strata-log.db
```

with:

```text
STRATA_LOG_STORAGE_PATH=/data/strata-log.db
```

A persistent volume prevents the database from being tied to the writable layer of a disposable container.

---

## Configuration

Configuration is provided through environment variables.

Important deployment settings include:

```text
STRATA_LOG_HOST
STRATA_LOG_PORT
STRATA_LOG_STORAGE_PATH
STRATA_LOG_STORAGE_RETRY_ATTEMPTS
STRATA_LOG_STORAGE_RETRY_BACKOFF
STRATA_LOG_BUFFER_CAPACITY
STRATA_LOG_BATCH_SIZE
STRATA_LOG_FLUSH_PERIOD
STRATA_LOG_SHUTDOWN_TIMEOUT
```

See:

```text
docs/configuration.md
```

for the complete configuration reference.

---

## Reverse Proxy

For deployments where Strata-Log is exposed beyond a trusted network, placing it behind a reverse proxy can provide capabilities such as:

* TLS termination
* authentication
* request filtering
* rate limiting
* access logging

Example:

```text
Internet
   │
   ▼
Reverse Proxy
   │
   ▼
Strata-Log :9090
   │
   ▼
SQLite
```

Strata-Log itself does not provide TLS termination or authentication in the current MVP.

---

## Backups

SQLite stores logs in a database file.

For important deployments:

* keep the database on persistent storage
* establish a backup process
* verify that backups can be restored
* retain more than one backup generation

Do not rely on a disposable container filesystem as the only copy of stored logs.

---

## Container Considerations

For production-like deployments, consider:

* running the container as a non-root user
* using persistent storage
* limiting container resources
* restricting network exposure
* terminating TLS at a reverse proxy
* monitoring `/healthz`
* scraping `/metrics`
* backing up the SQLite database

These are deployment concerns rather than features provided by the Strata-Log MVP itself.

---

## Graceful Shutdown

Strata-Log handles:

```text
SIGINT
SIGTERM
```

During shutdown, the application:

1. shuts down the HTTP server
2. stops accepting new processing work
3. drains pending pipeline work
4. flushes pending batches
5. closes SQLite
6. exits

The shutdown timeout is controlled by:

```text
STRATA_LOG_SHUTDOWN_TIMEOUT
```

---

## Deployment Verification

After starting the deployment, verify the following.

### 1. Container Status

```bash
docker compose \
  -f deployments/docker-compose.yml \
  ps
```

### 2. Health

```bash
curl http://localhost:9090/healthz
```

### 3. Metrics

```bash
curl http://localhost:9090/metrics
```

### 4. Ingestion

```bash
curl -X POST http://localhost:9090/v1/logs \
  -H 'Content-Type: application/json' \
  -d '{
    "timestamp": "2026-08-25T12:00:00Z",
    "level": "info",
    "service": "deployment",
    "message": "deployment verification"
  }'
```

### 5. Query

```bash
curl http://localhost:9090/v1/logs
```

---

## Troubleshooting

### Port 9090 Is Already in Use

Check which process is using the port:

```bash
ss -ltnp | grep 9090
```

Run Strata-Log on another port:

```bash
export STRATA_LOG_PORT=9191
```

---

### Database Permission Error

Check the configured path:

```bash
echo "$STRATA_LOG_STORAGE_PATH"
```

Ensure the process has permission to create and write the SQLite database.

For containers, verify that the mounted directory is writable by the container process.

---

### Container Cannot Access the Database

Verify:

1. the volume is mounted
2. the mount path matches `STRATA_LOG_STORAGE_PATH`
3. the directory exists
4. the container process can write to the directory

---

### Docker Cannot Connect to the Daemon

Check:

```bash
docker info
```

If Docker is not running, start the Docker engine and retry the deployment.

---

## Deployment Workflow

```text
Code
 │
 ▼
make check
 │
 ▼
make build
 │
 ▼
make docker-build
 │
 ▼
docker compose up
 │
 ▼
Health Check
 │
 ▼
API Verification
 │
 ▼
Metrics Verification
```

---

## Deployment Files

The container deployment is defined by:

```text
deployments/
├── Dockerfile
├── docker-compose.yml
└── prometheus.yml
```

# Strata-Log Deployment

This document describes how to build, run, test, and deploy Strata-Log.

---

## Requirements

### Local Development

* Go 1.26 or compatible version
* Git
* SQLite support provided through the Go SQLite driver

### Container Deployment

* Docker
* Docker Compose

---

## Run Locally

Clone the repository:

```bash
git clone https://github.com/ikwukao/strata-log.git
cd strata-log
```

Run the service:

```bash
go run ./cmd/strata-log
```

The default server listens on:

```text
http://localhost:9090
```

---

## Verify the Service

Health check:

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

---

## Query Logs

```bash
curl http://localhost:9090/v1/logs
```

---

## Metrics

```bash
curl http://localhost:9090/metrics
```

---

## Build

Build the binary:

```bash
go build -o bin/strata-log ./cmd/strata-log
```

Run it:

```bash
./bin/strata-log
```

---

## Test

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

Format the code:

```bash
gofmt -w .
```

---

## Makefile

The project provides a Makefile for common development operations.

Inspect available targets:

```bash
make help
```

Typical targets include:

```bash
make test
make test-race
make vet
make build
```

Use the Makefile where possible to keep local and CI workflows consistent.

---

## Docker

Build the container image:

```bash
docker compose -f deployments/docker-compose.yml build
```

Start Strata-Log:

```bash
docker compose -f deployments/docker-compose.yml up
```

Build and start:

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Run in the background:

```bash
docker compose -f deployments/docker-compose.yml up -d
```

Stop the deployment:

```bash
docker compose -f deployments/docker-compose.yml down
```

---

## Docker Health Check

Once the container is running:

```bash
curl http://localhost:9090/healthz
```

---

## Persistent Storage

The SQLite database should be stored on a persistent volume when running in containers.

Example:

```yaml
volumes:
  - strata-log-data:/data
```

The storage path should then point to:

```text
/data/strata-log.db
```

This prevents database data from being lost when the container is recreated.

---

## Production Deployment

A production deployment should provide:

* persistent SQLite storage
* controlled filesystem permissions
* container resource limits
* health checks
* metrics scraping
* log collection
* automated backups
* TLS termination where required
* restricted network access

Strata-Log itself should normally sit behind a reverse proxy or load balancer when exposed outside a trusted network.

---

## Reverse Proxy

A reverse proxy can provide:

* TLS termination
* authentication
* request filtering
* rate limiting
* access logging

Example architecture:

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

---

## Prometheus

The `/metrics` endpoint can be scraped by Prometheus.

Example:

```yaml
scrape_configs:
  - job_name: strata-log
    static_configs:
      - targets:
          - strata-log:9090
```

Prometheus will collect:

```text
strata_log_ingested_total
strata_log_stored_total
strata_log_errors_total
```

---

## Graceful Shutdown

Strata-Log handles:

```text
SIGINT
SIGTERM
```

During shutdown it:

1. stops accepting HTTP requests
2. shuts down the HTTP server
3. stops the processing pipeline
4. flushes pending batches
5. closes SQLite
6. exits

The shutdown duration is controlled by:

```text
STRATA_LOG_SHUTDOWN_TIMEOUT
```

---

## Database Backups

Because SQLite is file-based, the database file should be backed up regularly.

For production systems:

* use a persistent filesystem
* schedule backups
* verify backup integrity
* retain multiple backup generations

Do not store the only copy of production logs inside a disposable container filesystem.

---

## Container Security

The container should:

* run as a non-root user
* use a minimal base image
* expose only the required port
* avoid unnecessary Linux capabilities
* use read-only filesystem settings where practical
* mount only the required persistent storage

---

## Deployment Verification

After deployment, verify the following.

### 1. Container

```bash
docker compose ps
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

### Docker Cannot Connect to the Daemon

Check Docker:

```bash
docker info
```

If Docker Desktop is being used, ensure the Docker engine is running.

---

### Port Already in Use

Check which process is using port `9090`:

```bash
ss -ltnp | grep 9090
```

Change the port:

```bash
export STRATA_LOG_PORT=9191
```

---

### Database Permission Error

Check the configured storage path:

```bash
echo "$STRATA_LOG_STORAGE_PATH"
```

Ensure the process has read/write permissions.

---

### Container Cannot Access the Database

Verify that the SQLite directory is mounted as a persistent volume and that the configured path matches the mounted location.

---

## Recommended Deployment Flow

```text
Code
 │
 ▼
Run Tests
 │
 ▼
Race Detector
 │
 ▼
go vet
 │
 ▼
Build
 │
 ▼
Build Container
 │
 ▼
Start Container
 │
 ▼
Health Check
 │
 ▼
API Verification
 │
 ▼
Metrics Verification
 │
 ▼
Production
```

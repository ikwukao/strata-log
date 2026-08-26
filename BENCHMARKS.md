# Strata-Log Benchmarks

This document records the performance benchmarks for Strata-Log.

The goal is not to claim production-scale performance, but to provide reproducible reference measurements for the core ingestion pipeline and SQLite persistence layer.

---

## Benchmark Environment

The benchmark results documented here were collected on:

| Property      | Value                                    |
| ------------- | ---------------------------------------- |
| OS            | Linux                                    |
| Architecture  | amd64                                    |
| CPU           | Intel(R) Core(TM) i5-3427U CPU @ 1.80GHz |
| Runtime       | Go 1.26.4                                |
| Storage       | SQLite                                   |
| SQLite Driver | `modernc.org/sqlite`                     |

> Benchmark results are hardware- and workload-dependent. These numbers should be treated as local reference measurements rather than production performance guarantees.

---

## Running the Benchmarks

Run the complete benchmark suite:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

Run only the ingestion benchmarks:

```bash
go test -bench=. -benchmem ./benchmarks/ingest_test.go
```

Run only the storage benchmarks:

```bash
go test -bench=. -benchmem ./benchmarks/storage_test.go
```

Run a specific benchmark:

```bash
go test -bench=BenchmarkLogPipeline -benchmem ./benchmarks/...
```

Run benchmarks multiple times for more consistent measurements:

```bash
go test -bench=. -benchmem -count=5 ./benchmarks/...
```

---

## Benchmark Coverage

Strata-Log currently benchmarks two important parts of the system:

### 1. Log Ingestion Pipeline

Measures the overhead of processing structured log entries through the asynchronous ingestion pipeline.

The benchmark focuses on:

* Log entry processing
* Worker execution
* Channel-based coordination
* Callback dispatch
* Pipeline throughput
* Memory allocations

Benchmark:

```text
BenchmarkLogPipeline
```

### 2. SQLite Batch Persistence

Measures the cost of persisting a batch of log records to SQLite.

The benchmark covers:

* Database writes
* Batch persistence
* Transaction handling
* SQLite serialization
* Memory allocations

Benchmark:

```text
BenchmarkSQLiteWriteBatch
```

---

## Reference Results

Repeated local benchmark runs produced results in approximately these ranges:

| Benchmark                   | Result           |
| --------------------------- | ---------------- |
| `BenchmarkLogPipeline`      | ~1.03–1.22 µs/op |
| `BenchmarkLogPipeline`      | 0 B/op           |
| `BenchmarkLogPipeline`      | 0 allocs/op      |
| `BenchmarkSQLiteWriteBatch` | ~10.1–11.9 ms/op |
| `BenchmarkSQLiteWriteBatch` | ~63 KB/op        |
| `BenchmarkSQLiteWriteBatch` | ~2,012 allocs/op |

Example run:

```text
BenchmarkLogPipeline-2
    1000000    1139 ns/op    0 B/op    0 allocs/op

BenchmarkSQLiteWriteBatch-2
    122    9379326 ns/op    63070 B/op    2012 allocs/op
```

---

## Interpreting the Results

### Ingestion Pipeline

The ingestion benchmark demonstrates that the in-memory processing path can operate with very low overhead.

The reference results show:

```text
~1 µs/op
0 B/op
0 allocs/op
```

The zero-allocation result is particularly useful because the ingestion path can execute at high frequency without introducing additional heap allocation pressure in the benchmarked code path.

This does **not** mean the complete HTTP ingestion request is zero-allocation. HTTP parsing, JSON decoding, request handling, and other components introduce their own allocations.

---

### SQLite Persistence

SQLite persistence is substantially more expensive than the in-memory ingestion pipeline.

This is expected because the storage benchmark includes database operations and transaction-related work.

The reference measurements were approximately:

```text
~10–12 ms/op
~63 KB/op
~2,000 allocs/op
```

The storage path is therefore expected to be the more expensive part of the system.

This is one reason Strata-Log uses:

```text
HTTP
  ↓
Ingestion
  ↓
Asynchronous Processing
  ↓
Batcher
  ↓
SQLite
```

rather than performing a database write directly inside the HTTP request path.

---

## Why Batching Matters

Writing every log entry individually would introduce database overhead for every event.

Strata-Log instead accumulates entries and writes them in batches.

Conceptually:

```text
Individual Writes

Log 1 ──→ SQLite
Log 2 ──→ SQLite
Log 3 ──→ SQLite
Log 4 ──→ SQLite
```

versus:

```text
Batching

Log 1 ─┐
Log 2 ─┤
Log 3 ─┤──→ Batch ──→ SQLite
Log 4 ─┘
```

Batching reduces the relative cost of persistence operations and allows the ingestion path to remain decoupled from database latency.

---

## Benchmarking Principles

Performance changes should be measured rather than assumed.

When making a performance-related change:

1. Establish a baseline.
2. Run the relevant benchmark.
3. Make one change.
4. Run the benchmark again.
5. Compare the results.
6. Verify correctness with the test suite.
7. Run the race detector.
8. Run `go vet`.

Recommended validation:

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
go test -bench=. -benchmem ./benchmarks/...
```

---

## Benchmark Limitations

These benchmarks do not represent a production deployment.

Results can vary significantly based on:

* CPU
* RAM
* Operating system
* Filesystem
* SQLite configuration
* Database size
* Batch size
* Workload characteristics
* Go version
* Compiler optimizations
* System load
* Containerization
* Storage hardware

The benchmark suite should therefore be used primarily for **relative comparisons between implementations**, not as an absolute capacity guarantee.

For example, a change that improves:

```text
1.20 µs/op → 0.90 µs/op
```

on the same machine is meaningful even though the absolute value will differ on another system.

---

## Future Benchmark Areas

As Strata-Log evolves, additional benchmarks may cover:

* HTTP ingestion throughput
* JSON decoding
* Concurrent ingestion
* Different batch sizes
* SQLite read performance
* Query filtering
* Query pagination
* Storage retry behavior
* Memory usage under sustained load
* Large log fields
* High-concurrency workloads

These should be added only when they provide useful engineering information.

---

## Reproducibility

Benchmark source code is located in:

```text
benchmarks/
├── ingest_test.go
└── storage_test.go
```

Run:

```bash
go test -bench=. -benchmem ./benchmarks/...
```

to reproduce the measurements on your own system.

---

## Summary

Strata-Log separates fast in-memory ingestion from comparatively expensive persistent storage.

The current benchmark results demonstrate:

```text
Fast ingestion path
        ↓
Asynchronous processing
        ↓
Batching
        ↓
Durable SQLite persistence
```

The benchmarks support the architectural decision to keep database operations away from the synchronous ingestion path while using batching to improve persistence efficiency.

Performance should continue to be evaluated empirically as the system evolves.

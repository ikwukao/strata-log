// Package storage defines persistence interfaces and storage backends
// used by Strata-Log.
package storage

import (
	"context"

	"github.com/ikwukao/strata-log/internal/ingest"
)

// Writer persists batches of log entries.
type Writer interface {
	WriteBatch(context.Context, []ingest.LogEntry) error
}

// Reader retrieves persisted log entries.
type Reader interface {
	QueryLogs(context.Context, QueryOptions) ([]LogRecord, error)
}

// QueryOptions controls log retrieval.
type QueryOptions struct {
	Level   string
	Service string
	Limit   int
}

// LogRecord represents a persisted log entry.
type LogRecord struct {
	ID        int64
	Timestamp string
	Level     string
	Service   string
	Message   string
	Fields    map[string]any
}

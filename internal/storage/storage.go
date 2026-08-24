// Package storage defines persistence interfaces and storage backends
// used by Strata-Log.
package storage

import (
	"context"
	"time"

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
//
// Filters are combined using AND semantics. Results are returned
// newest-first and constrained by Limit and Offset.
type QueryOptions struct {
	// Level filters logs by severity level.
	Level string

	// Service filters logs by originating service.
	Service string

	// From filters out logs older than this timestamp.
	From *time.Time

	// To filters out logs newer than this timestamp.
	To *time.Time

	// Limit controls the maximum number of records returned.
	Limit int

	// Offset skips the first Offset matching records.
	Offset int
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

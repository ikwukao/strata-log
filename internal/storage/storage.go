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

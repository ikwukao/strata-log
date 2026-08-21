package storage

import (
	"context"
	"sync"

	"github.com/ikwukao/strata-log/internal/ingest"
)

// MemoryWriter stores log entries in memory.
//
// It is primarily useful for tests and local development.
type MemoryWriter struct {
	mu      sync.RWMutex
	entries []ingest.LogEntry
}

// NewMemoryWriter creates an empty in-memory writer.
func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{
		entries: make([]ingest.LogEntry, 0),
	}
}

// WriteBatch appends a batch of log entries to memory.
func (w *MemoryWriter) WriteBatch(
	ctx context.Context,
	entries []ingest.LogEntry,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.entries = append(w.entries, entries...)

	return nil
}

// Entries returns a snapshot of all stored log entries.
func (w *MemoryWriter) Entries() []ingest.LogEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entries := make([]ingest.LogEntry, len(w.entries))
	copy(entries, w.entries)

	return entries
}

// Len returns the number of stored log entries.
func (w *MemoryWriter) Len() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return len(w.entries)
}

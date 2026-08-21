package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/ikwukao/strata-log/internal/ingest"
)

// FileWriter persists log entries as newline-delimited JSON.
//
// Entries are appended sequentially to a single file. The writer owns
// the file handle and serializes concurrent writes.
type FileWriter struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileWriter opens or creates an append-only log file.
//
// The parent directory is created automatically when necessary.
func NewFileWriter(path string) (*FileWriter, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file: file,
	}, nil
}

// WriteBatch appends a batch of log entries to disk.
func (w *FileWriter) WriteBatch(
	ctx context.Context,
	entries []ingest.LogEntry,
) error {
	if len(entries) == 0 {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	writer := bufio.NewWriter(w.file)

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		if _, err := writer.Write(data); err != nil {
			return err
		}

		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return w.file.Sync()
}

// Close closes the underlying storage file.
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil

	return err
}

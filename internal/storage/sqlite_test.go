package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
)

func TestSQLiteWriterPersistsBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	entries := []ingest.LogEntry{
		storageTestEntry("first"),
		storageTestEntry("second"),
		storageTestEntry("third"),
	}

	if err := writer.WriteBatch(
		context.Background(),
		entries,
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	var count int

	if err := writer.db.QueryRow(
		"SELECT COUNT(*) FROM logs",
	).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}

	if count != 3 {
		t.Fatalf("expected 3 logs, got %d", count)
	}
}

func TestSQLiteWriterPersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	if err := writer.WriteBatch(
		context.Background(),
		[]ingest.LogEntry{
			storageTestEntry("persistent log"),
		},
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	writer, err = NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}

	defer writer.Close()

	var message string

	err = writer.db.QueryRow(
		"SELECT message FROM logs LIMIT 1",
	).Scan(&message)

	if err != nil {
		t.Fatalf("query persisted log: %v", err)
	}

	if message != "persistent log" {
		t.Fatalf(
			"expected persisted message %q, got %q",
			"persistent log",
			message,
		)
	}
}

func TestSQLiteWriterUsesWAL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	var mode string

	if err := writer.db.QueryRow(
		"PRAGMA journal_mode",
	).Scan(&mode); err != nil {
		t.Fatalf("journal mode query failed: %v", err)
	}

	if mode != "wal" {
		t.Fatalf(
			"expected WAL journal mode, got %q",
			mode,
		)
	}
}

func TestSQLiteWriterEmptyBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	if err := writer.WriteBatch(
		context.Background(),
		nil,
	); err != nil {
		t.Fatalf("empty WriteBatch() error = %v", err)
	}

	var count int

	if err := writer.db.QueryRow(
		"SELECT COUNT(*) FROM logs",
	).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected 0 logs, got %d", count)
	}
}

func TestSQLiteWriterContextCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = writer.WriteBatch(
		ctx,
		[]ingest.LogEntry{
			storageTestEntry("cancelled"),
		},
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

var (
	_ Writer = (*SQLiteWriter)(nil)
	_ Reader = (*SQLiteWriter)(nil)
)

// Atomicity Test
func TestSQLiteWriterBatchIsAtomic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	count := 0

	writer.insertHook = func(entry ingest.LogEntry) error {
		count++

		if count == 2 {
			return errors.New("simulated insert failure")
		}

		return nil
	}

	entries := []ingest.LogEntry{
		storageTestEntry("first"),
		storageTestEntry("second"),
		storageTestEntry("third"),
	}

	err = writer.WriteBatch(
		context.Background(),
		entries,
	)

	if err == nil {
		t.Fatal("expected WriteBatch() to fail")
	}

	var stored int

	if err := writer.db.QueryRow(
		"SELECT COUNT(*) FROM logs",
	).Scan(&stored); err != nil {
		t.Fatalf("count query failed: %v", err)
	}

	if stored != 0 {
		t.Fatalf(
			"expected transaction rollback to leave 0 logs, got %d",
			stored,
		)
	}
}

// Storage Query Test
func TestSQLiteWriterQueryLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	entries := []ingest.LogEntry{
		{
			Timestamp: time.Date(
				2026,
				time.August,
				22,
				20,
				0,
				0,
				0,
				time.UTC,
			),
			Level:   "info",
			Service: "api",
			Message: "request accepted",
			Fields: map[string]any{
				"request_id": "req-1",
				"status":     200,
			},
		},
		{
			Timestamp: time.Date(
				2026,
				time.August,
				22,
				21,
				0,
				0,
				0,
				time.UTC,
			),
			Level:   "error",
			Service: "api",
			Message: "database unavailable",
			Fields: map[string]any{
				"request_id": "req-2",
				"status":     500,
			},
		},
	}

	if err := writer.WriteBatch(
		context.Background(),
		entries,
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	records, err := writer.QueryLogs(
		context.Background(),
		QueryOptions{
			Limit: 10,
		},
	)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	if records[0].Message != "database unavailable" {
		t.Fatalf(
			"expected newest record first, got %q",
			records[0].Message,
		)
	}
}

// Filter Coverage
func TestSQLiteWriterQueryFilters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := NewSQLiteWriter(path)
	if err != nil {
		t.Fatalf("NewSQLiteWriter() error = %v", err)
	}

	defer writer.Close()

	entries := []ingest.LogEntry{
		storageTestEntryWithService(
			"error from payments",
			"payment-api",
			"error",
		),
		storageTestEntryWithService(
			"info from payments",
			"payment-api",
			"info",
		),
		storageTestEntryWithService(
			"error from users",
			"user-api",
			"error",
		),
	}

	if err := writer.WriteBatch(
		context.Background(),
		entries,
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	records, err := writer.QueryLogs(
		context.Background(),
		QueryOptions{
			Level:   "error",
			Service: "payment-api",
			Limit:   10,
		},
	)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf(
			"expected 1 filtered record, got %d",
			len(records),
		)
	}

	if records[0].Message != "error from payments" {
		t.Fatalf(
			"unexpected record: %q",
			records[0].Message,
		)
	}
}

func storageTestEntryWithService(
	message string,
	service string,
	level string,
) ingest.LogEntry {
	return ingest.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Service:   service,
		Message:   message,
		Fields: map[string]any{
			"test": true,
		},
	}
}

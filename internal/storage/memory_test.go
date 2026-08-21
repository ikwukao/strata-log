package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
)

func storageTestEntry(message string) ingest.LogEntry {
	return ingest.LogEntry{
		Timestamp: time.Date(
			2026,
			time.August,
			21,
			18,
			0,
			0,
			0,
			time.UTC,
		),
		Level:   "info",
		Service: "test-service",
		Message: message,
	}
}

func TestMemoryWriterWriteBatch(t *testing.T) {
	writer := NewMemoryWriter()

	entries := []ingest.LogEntry{
		storageTestEntry("first"),
		storageTestEntry("second"),
		storageTestEntry("third"),
	}

	if err := writer.WriteBatch(context.Background(), entries); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	if got := writer.Len(); got != 3 {
		t.Fatalf("expected 3 entries, got %d", got)
	}

	stored := writer.Entries()

	if len(stored) != len(entries) {
		t.Fatalf(
			"expected %d entries, got %d",
			len(entries),
			len(stored),
		)
	}

	for i := range entries {
		if stored[i].Message != entries[i].Message {
			t.Errorf(
				"entry %d: expected message %q, got %q",
				i,
				entries[i].Message,
				stored[i].Message,
			)
		}
	}
}

func TestMemoryWriterPreservesBatchOrder(t *testing.T) {
	writer := NewMemoryWriter()

	entries := []ingest.LogEntry{
		storageTestEntry("first"),
		storageTestEntry("second"),
		storageTestEntry("third"),
	}

	if err := writer.WriteBatch(context.Background(), entries); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	stored := writer.Entries()

	for i, want := range []string{
		"first",
		"second",
		"third",
	} {
		if stored[i].Message != want {
			t.Errorf(
				"entry %d: expected %q, got %q",
				i,
				want,
				stored[i].Message,
			)
		}
	}
}

func TestMemoryWriterMultipleBatches(t *testing.T) {
	writer := NewMemoryWriter()

	first := []ingest.LogEntry{
		storageTestEntry("first"),
		storageTestEntry("second"),
	}

	second := []ingest.LogEntry{
		storageTestEntry("third"),
		storageTestEntry("fourth"),
	}

	if err := writer.WriteBatch(context.Background(), first); err != nil {
		t.Fatalf("first WriteBatch() error = %v", err)
	}

	if err := writer.WriteBatch(context.Background(), second); err != nil {
		t.Fatalf("second WriteBatch() error = %v", err)
	}

	if got := writer.Len(); got != 4 {
		t.Fatalf("expected 4 entries, got %d", got)
	}

	stored := writer.Entries()

	want := []string{
		"first",
		"second",
		"third",
		"fourth",
	}

	for i, message := range want {
		if stored[i].Message != message {
			t.Errorf(
				"entry %d: expected %q, got %q",
				i,
				message,
				stored[i].Message,
			)
		}
	}
}

func TestMemoryWriterContextCancellation(t *testing.T) {
	writer := NewMemoryWriter()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := writer.WriteBatch(
		ctx,
		[]ingest.LogEntry{
			storageTestEntry("should not be stored"),
		},
	)

	if err != context.Canceled {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if got := writer.Len(); got != 0 {
		t.Fatalf(
			"expected 0 entries after cancelled write, got %d",
			got,
		)
	}
}

func TestMemoryWriterEntriesReturnsCopy(t *testing.T) {
	writer := NewMemoryWriter()

	if err := writer.WriteBatch(
		context.Background(),
		[]ingest.LogEntry{
			storageTestEntry("original"),
		},
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	entries := writer.Entries()

	entries[0].Message = "modified"

	stored := writer.Entries()

	if stored[0].Message != "original" {
		t.Fatalf(
			"modifying returned entries changed internal storage: got %q",
			stored[0].Message,
		)
	}
}

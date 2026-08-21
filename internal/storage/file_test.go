package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
)

func TestFileWriterWritesBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata.log")

	writer, err := NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter() error = %v", err)
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

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var stored []ingest.LogEntry

	for scanner.Scan() {
		var entry ingest.LogEntry

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON line: %v", err)
		}

		stored = append(stored, entry)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

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
				"entry %d: expected %q, got %q",
				i,
				entries[i].Message,
				stored[i].Message,
			)
		}
	}
}

func TestFileWriterAppendsBatches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata.log")

	writer, err := NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter() error = %v", err)
	}

	first := []ingest.LogEntry{
		storageTestEntry("first"),
	}

	second := []ingest.LogEntry{
		storageTestEntry("second"),
	}

	if err := writer.WriteBatch(context.Background(), first); err != nil {
		t.Fatalf("first WriteBatch() error = %v", err)
	}

	if err := writer.WriteBatch(context.Background(), second); err != nil {
		t.Fatalf("second WriteBatch() error = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var messages []string

	for scanner.Scan() {
		var entry ingest.LogEntry

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON line: %v", err)
		}

		messages = append(messages, entry.Message)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(messages))
	}

	if messages[0] != "first" {
		t.Errorf("expected first entry to be first, got %q", messages[0])
	}

	if messages[1] != "second" {
		t.Errorf("expected second entry to be second, got %q", messages[1])
	}
}

func TestFileWriterCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"nested",
		"logs",
		"strata.log",
	)

	writer, err := NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter() error = %v", err)
	}

	defer writer.Close()

	entry := ingest.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Service:   "test",
		Message:   "directory creation works",
	}

	if err := writer.WriteBatch(
		context.Background(),
		[]ingest.LogEntry{entry},
	); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
}

func TestFileWriterContextCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strata.log")

	writer, err := NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter() error = %v", err)
	}

	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = writer.WriteBatch(
		ctx,
		[]ingest.LogEntry{
			storageTestEntry("should not be written"),
		},
	)

	if err != context.Canceled {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

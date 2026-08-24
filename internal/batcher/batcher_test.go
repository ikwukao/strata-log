package batcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/storage"
)

func testEntry(message string) ingest.LogEntry {
	return ingest.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Service:   "test",
		Message:   message,
	}
}

func TestBatcherFlushesAtBatchSize(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		3,
		time.Hour,
		10,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	defer b.Close()

	for i := 0; i < 3; i++ {
		if err := b.Submit(testEntry("message")); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	deadline := time.After(time.Second)

	for {
		if writer.Len() == 3 {
			break
		}

		select {
		case <-deadline:
			t.Fatalf(
				"expected 3 stored entries, got %d",
				writer.Len(),
			)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func TestBatcherFlushesOnClose(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		100,
		time.Hour,
		10,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := b.Submit(testEntry("message")); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	b.Close()

	if got := writer.Len(); got != 5 {
		t.Fatalf("expected 5 stored entries, got %d", got)
	}
}

func TestBatcherFlushesOnInterval(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		100,
		10*time.Millisecond,
		10,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	defer b.Close()

	if err := b.Submit(testEntry("message")); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	deadline := time.After(time.Second)

	for {
		if writer.Len() == 1 {
			break
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for batch flush")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

type failingWriter struct {
	err error
}

func (w *failingWriter) WriteBatch(
	context.Context,
	[]ingest.LogEntry,
) error {
	return w.err
}

// Test storage failures
func TestBatcherReportsStorageErrors(t *testing.T) {
	expectedErr := errors.New("storage unavailable")

	writer := &failingWriter{
		err: expectedErr,
	}

	b, err := New(
		context.Background(),
		writer,
		1,
		time.Hour,
		10,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	defer b.Close()

	if err := b.Submit(testEntry("message")); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case err := <-b.Errors():
		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected error %v, got %v",
				expectedErr,
				err,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for storage error")
	}
}

func TestBatcherReportsStoredEntries(t *testing.T) {
	writer := storage.NewMemoryWriter()

	ctx := context.Background()

	b, err := New(
		ctx,
		writer,
		2,
		time.Hour,
		10,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := b.Submit(ingest.LogEntry{
		Message: "first",
	}); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if err := b.Submit(ingest.LogEntry{
		Message: "second",
	}); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case count := <-b.Stored():
		if count != 2 {
			t.Fatalf(
				"expected 2 stored entries, got %d",
				count,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stored notification")
	}

	b.Close()
}

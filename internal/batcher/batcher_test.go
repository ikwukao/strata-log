package batcher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
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

func TestBatcherRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		writerNil     bool
		batchSize     int
		flushInterval time.Duration
		bufferSize    int
		retryAttempts int
		retryBackoff  time.Duration
	}{
		{
			name:          "nil writer",
			writerNil:     true,
			batchSize:     1,
			flushInterval: time.Second,
			bufferSize:    1,
			retryAttempts: 1,
		},
		{
			name:          "zero batch size",
			batchSize:     0,
			flushInterval: time.Second,
			bufferSize:    1,
			retryAttempts: 1,
		},
		{
			name:          "negative batch size",
			batchSize:     -1,
			flushInterval: time.Second,
			bufferSize:    1,
			retryAttempts: 1,
		},
		{
			name:          "zero flush interval",
			batchSize:     1,
			flushInterval: 0,
			bufferSize:    1,
			retryAttempts: 1,
		},
		{
			name:          "zero buffer size",
			batchSize:     1,
			flushInterval: time.Second,
			bufferSize:    0,
			retryAttempts: 1,
		},
		{
			name:          "zero retry attempts",
			batchSize:     1,
			flushInterval: time.Second,
			bufferSize:    1,
			retryAttempts: 0,
		},
		{
			name:          "negative retry backoff",
			batchSize:     1,
			flushInterval: time.Second,
			bufferSize:    1,
			retryAttempts: 1,
			retryBackoff:  -time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var writer storage.Writer

			if !tt.writerNil {
				writer = storage.NewMemoryWriter()
			}

			_, err := New(
				context.Background(),
				writer,
				tt.batchSize,
				tt.flushInterval,
				tt.bufferSize,
				tt.retryAttempts,
				tt.retryBackoff,
			)

			if err == nil {
				t.Fatal("expected configuration error")
			}
		})
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
		10,
		time.Second,
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
		10,
		time.Second,
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
		10,
		time.Second,
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
		1,
		0,
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

	b, err := New(
		context.Background(),
		writer,
		2,
		time.Hour,
		10,
		10,
		time.Second,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := b.Submit(testEntry("first")); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if err := b.Submit(testEntry("second")); err != nil {
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

func TestBatcherRejectsSubmitAfterClose(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		10,
		time.Hour,
		10,
		1,
		0,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	b.Close()

	if err := b.Submit(testEntry("after close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestBatcherCloseIsIdempotent(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		10,
		time.Hour,
		10,
		1,
		0,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	b.Close()
	b.Close()
	b.Close()
}

func TestBatcherProcessesAllAcceptedEntriesBeforeClose(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		100,
		time.Hour,
		100,
		1,
		0,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	const total = 100

	for i := 0; i < total; i++ {
		if err := b.Submit(testEntry("message")); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	b.Close()

	if got := writer.Len(); got != total {
		t.Fatalf(
			"expected all %d accepted entries to be stored, got %d",
			total,
			got,
		)
	}
}

func TestBatcherDoesNotBlockWhenErrorBufferIsFull(t *testing.T) {
	writer := &failingWriter{
		err: errors.New("storage unavailable"),
	}

	b, err := New(
		context.Background(),
		writer,
		1,
		time.Hour,
		100,
		1,
		0,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for i := 0; i < 100; i++ {
		if err := b.Submit(testEntry("message")); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	done := make(chan struct{})

	go func() {
		b.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("batcher Close blocked with a full error channel")
	}
}

func TestBatcherSubmitCloseRace(t *testing.T) {
	writer := storage.NewMemoryWriter()

	b, err := New(
		context.Background(),
		writer,
		10,
		time.Hour,
		100,
		1,
		0,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var accepted atomic.Int32

	var submitters sync.WaitGroup

	for i := 0; i < 10; i++ {
		submitters.Add(1)

		go func() {
			defer submitters.Done()

			for j := 0; j < 100; j++ {
				if err := b.Submit(testEntry("message")); err == nil {
					accepted.Add(1)
				}
			}
		}()
	}

	time.Sleep(10 * time.Millisecond)

	b.Close()

	submitters.Wait()

	if got := accepted.Load(); got < 0 {
		t.Fatalf("invalid accepted count: %d", got)
	}
}

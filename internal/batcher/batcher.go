package batcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/storage"
)

var (
	// ErrClosed indicates that the batcher is no longer accepting entries.
	ErrClosed = errors.New("batcher is closed")
)

// Batcher collects log entries and periodically writes them as batches.
type Batcher struct {
	writer storage.Writer

	input chan ingest.LogEntry

	batchSize     int
	flushInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	errors chan error

	wg sync.WaitGroup

	closeOnce sync.Once
}

// New creates and starts a Batcher.
func New(
	parent context.Context,
	writer storage.Writer,
	batchSize int,
	flushInterval time.Duration,
	bufferSize int,
) (*Batcher, error) {
	if writer == nil {
		return nil, errors.New("storage writer is required")
	}

	if batchSize <= 0 {
		return nil, errors.New("batch size must be greater than zero")
	}

	if flushInterval <= 0 {
		return nil, errors.New("flush interval must be greater than zero")
	}

	if bufferSize <= 0 {
		return nil, errors.New("buffer size must be greater than zero")
	}

	ctx, cancel := context.WithCancel(parent)

	b := &Batcher{
		writer:        writer,
		errors:        make(chan error, 16),
		input:         make(chan ingest.LogEntry, bufferSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		ctx:           ctx,
		cancel:        cancel,
	}

	b.wg.Add(1)

	go b.run()

	return b, nil
}

// Errors returns a channel containing asynchronous storage errors.
func (b *Batcher) Errors() <-chan error {
	return b.errors
}

// Submit adds a log entry to the batcher.
func (b *Batcher) Submit(entry ingest.LogEntry) error {
	select {
	case <-b.ctx.Done():
		return ErrClosed

	case b.input <- entry:
		return nil
	}
}

// Close stops accepting new entries and flushes remaining entries.
func (b *Batcher) Close() {
	b.closeOnce.Do(func() {
		close(b.input)
		b.wg.Wait()
		b.cancel()
	})
}

func (b *Batcher) run() {
	defer b.wg.Done()
	defer close(b.errors)

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([]ingest.LogEntry, 0, b.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		entries := make([]ingest.LogEntry, len(batch))
		copy(entries, batch)

		if err := b.writer.WriteBatch(b.ctx, entries); err != nil {
			// Storage error handling will be connected to resilience
			// and telemetry in a later milestone.
			select {
			case b.errors <- err:
			case <-b.ctx.Done():
			}

			return
		}

		batch = batch[:0]
	}

	for {
		select {
		case entry, ok := <-b.input:
			if !ok {
				flush()
				return
			}

			batch = append(batch, entry)

			if len(batch) >= b.batchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-b.ctx.Done():
			return
		}
	}
}

// Package batcher provides buffered batching and persistence for log entries.
package batcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/resilience"
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
	retryAttempts int
	retryBackoff  time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	errors chan error
	stored chan int

	wg sync.WaitGroup

	mu     sync.RWMutex
	closed bool
}

// New creates and starts a Batcher.
func New(
	parent context.Context,
	writer storage.Writer,
	batchSize int,
	flushInterval time.Duration,
	bufferSize int,
	retryAttempts int,
	retryBackoff time.Duration,
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

	if retryAttempts <= 0 {
		return nil, errors.New("retry attempts must be greater than zero")
	}

	if retryBackoff < 0 {
		return nil, errors.New("retry backoff must not be negative")
	}

	ctx, cancel := context.WithCancel(parent)

	b := &Batcher{
		writer:        writer,
		input:         make(chan ingest.LogEntry, bufferSize),
		errors:        make(chan error, 16),
		stored:        make(chan int, 16),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		retryAttempts: retryAttempts,
		retryBackoff:  retryBackoff,
		ctx:           ctx,
		cancel:        cancel,
	}

	b.wg.Add(1)
	go b.run()

	return b, nil
}

// Errors returns a channel containing asynchronous storage errors.
//
// Errors are delivered on a best-effort basis. A full error channel
// never blocks the batcher from continuing or shutting down.
func (b *Batcher) Errors() <-chan error {
	return b.errors
}

// Stored returns a channel containing the number of entries
// successfully persisted by each batch.
func (b *Batcher) Stored() <-chan int {
	return b.stored
}

// Submit adds a log entry to the batcher.
//
// Submit is safe to call concurrently with Close. Once Close begins,
// new submissions are rejected with ErrClosed.
func (b *Batcher) Submit(entry ingest.LogEntry) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrClosed
	}

	select {
	case <-b.ctx.Done():
		return ErrClosed

	case b.input <- entry:
		return nil
	}
}

// Close stops accepting new entries, drains all accepted entries,
// waits for the batcher to stop, and releases its resources.
//
// Close is safe to call multiple times.
func (b *Batcher) Close() {
	b.mu.Lock()

	if b.closed {
		b.mu.Unlock()
		return
	}

	b.closed = true
	close(b.input)

	b.mu.Unlock()

	// The run loop drains the input channel and flushes the final batch
	// before returning.
	b.wg.Wait()

	b.cancel()
}

// run owns the input channel and performs all batching and persistence.
func (b *Batcher) run() {
	defer b.wg.Done()
	defer close(b.errors)
	defer close(b.stored)

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([]ingest.LogEntry, 0, b.batchSize)

	flush := func() bool {
		if len(batch) == 0 {
			return true
		}

		entries := make([]ingest.LogEntry, len(batch))
		copy(entries, batch)

		err := resilience.Retry(
			b.ctx,
			b.retryAttempts,
			b.retryBackoff,
			func(ctx context.Context) error {
				return b.writer.WriteBatch(ctx, entries)
			},
		)

		if err != nil {
			// Shutdown-related cancellation is not a storage failure.
			if !errors.Is(err, context.Canceled) &&
				!errors.Is(err, context.DeadlineExceeded) {
				b.reportError(err)
			}

			return false
		}

		// A successful write should always be reported when possible.
		// Never allow the notification channel to block persistence.
		b.reportStored(len(entries))

		batch = batch[:0]

		return true
	}

	for {
		select {
		case entry, ok := <-b.input:
			if !ok {
				// Close drains all entries already accepted by Submit.
				flush()
				return
			}

			batch = append(batch, entry)

			if len(batch) >= b.batchSize {
				if !flush() {
					return
				}
			}

		case <-ticker.C:
			if !flush() {
				return
			}

		case <-b.ctx.Done():
			return
		}
	}
}

// reportError publishes a storage error without blocking the batcher.
func (b *Batcher) reportError(err error) {
	select {
	case b.errors <- err:
	default:
	}
}

// reportStored publishes a successful batch notification without blocking.
func (b *Batcher) reportStored(count int) {
	select {
	case b.stored <- count:
	default:
	}
}

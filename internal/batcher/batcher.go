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

	closeOnce sync.Once
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
func (b *Batcher) Errors() <-chan error {
	return b.errors
}

// Stored returns a channel containing the number of entries
// successfully persisted by each batch.
func (b *Batcher) Stored() <-chan int {
	return b.stored
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

// Close stops accepting new entries and waits for the batcher to stop.
//
// Cancellation is triggered first so any in-flight retry operation can
// terminate promptly. The input channel is then closed so no new entries
// can be submitted.
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
			// A canceled context means shutdown was requested. Do not
			// report cancellation as a storage failure.
			if !errors.Is(err, context.Canceled) &&
				!errors.Is(err, context.DeadlineExceeded) {
				select {
				case b.errors <- err:
				case <-b.ctx.Done():
				}
			}

			return false
		}

		select {
		case b.stored <- len(entries):
		case <-b.ctx.Done():
			return false
		}

		batch = batch[:0]

		return true
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

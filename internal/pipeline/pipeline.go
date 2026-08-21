package pipeline

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("pipeline is closed")

type Handler[T any] func(context.Context, T) error

type Pipeline[T any] struct {
	handler Handler[T]

	input  chan T
	errors chan error

	workers int

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	mu     sync.RWMutex
	closed bool
}

func New[T any](
	parent context.Context,
	workers int,
	bufferSize int,
	handler Handler[T],
) (*Pipeline[T], error) {
	if workers <= 0 {
		return nil, errors.New("worker count must be greater than zero")
	}

	if bufferSize <= 0 {
		return nil, errors.New("buffer size must be greater than zero")
	}

	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(parent)

	p := &Pipeline[T]{
		handler: handler,
		input:   make(chan T, bufferSize),
		errors:  make(chan error, bufferSize),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}

	p.start()

	return p, nil
}

func (p *Pipeline[T]) start() {
	p.wg.Add(p.workers)

	for i := 0; i < p.workers; i++ {
		go p.worker()
	}
}

func (p *Pipeline[T]) worker() {
	defer p.wg.Done()

	for item := range p.input {
		if err := p.handler(p.ctx, item); err != nil {
			select {
			case p.errors <- err:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

func (p *Pipeline[T]) Submit(item T) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrClosed
	}

	select {
	case <-p.ctx.Done():
		return ErrClosed

	case p.input <- item:
		return nil
	}
}

func (p *Pipeline[T]) Errors() <-chan error {
	return p.errors
}

func (p *Pipeline[T]) Close() {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return
	}

	p.closed = true
	p.mu.Unlock()

	// Stop accepting new work, but allow workers to
	// drain everything already queued in input.
	close(p.input)

	// Wait until all queued work has been processed.
	p.wg.Wait()

	// Cancel the shared context only after the queue
	// has been completely drained.
	p.cancel()

	close(p.errors)
}

func (p *Pipeline[T]) Wait() {
	p.wg.Wait()
}

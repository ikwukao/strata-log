// Package buffer provides bounded in-memory buffering with explicit
// capacity and lifecycle semantics.
package buffer

import (
	"errors"
	"sync"
)

var (
	// ErrFull indicates that the buffer has reached its configured capacity.
	ErrFull = errors.New("buffer is full")

	// ErrClosed indicates that the buffer no longer accepts new items.
	ErrClosed = errors.New("buffer is closed")
)

// Buffer is a bounded, thread-safe FIFO buffer.
//
// A Buffer rejects new items when its configured capacity is reached.
// Once closed, the buffer cannot accept additional items.
type Buffer[T any] struct {
	mu     sync.RWMutex
	items  []T
	cap    int
	closed bool
}

// New creates a Buffer with the specified maximum capacity.
func New[T any](capacity int) (*Buffer[T], error) {
	if capacity <= 0 {
		return nil, errors.New("buffer capacity must be greater than zero")
	}

	return &Buffer[T]{
		items: make([]T, 0, capacity),
		cap:   capacity,
	}, nil
}

// Push adds an item to the buffer.
//
// Push returns ErrFull when the buffer has reached its capacity and
// ErrClosed when the buffer has been closed.
func (b *Buffer[T]) Push(item T) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrClosed
	}

	if len(b.items) >= b.cap {
		return ErrFull
	}

	b.items = append(b.items, item)

	return nil
}

// Pop removes and returns the oldest item in the buffer.
//
// Pop returns false when the buffer is empty.
func (b *Buffer[T]) Pop() (T, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.items) == 0 {
		var zero T
		return zero, false
	}

	item := b.items[0]

	var zero T
	b.items[0] = zero

	b.items = b.items[1:]

	return item, true
}

// Len returns the number of items currently stored in the buffer.
func (b *Buffer[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.items)
}

// Capacity returns the maximum number of items the buffer can hold.
func (b *Buffer[T]) Capacity() int {
	return b.cap
}

// Close prevents the buffer from accepting additional items.
func (b *Buffer[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
}

// Closed reports whether the buffer has been closed.
func (b *Buffer[T]) Closed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.closed
}

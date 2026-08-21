package buffer

import (
	"errors"
	"sync"
)

var (
	ErrFull   = errors.New("buffer is full")
	ErrClosed = errors.New("buffer is closed")
)

type Buffer[T any] struct {
	mu     sync.RWMutex
	items  []T
	cap    int
	closed bool
}

func New[T any](capacity int) (*Buffer[T], error) {
	if capacity <= 0 {
		return nil, errors.New("buffer capacity must be greater than zero")
	}

	return &Buffer[T]{
		items: make([]T, 0, capacity),
		cap:   capacity,
	}, nil
}

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

func (b *Buffer[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.items)
}

func (b *Buffer[T]) Capacity() int {
	return b.cap
}

func (b *Buffer[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
}

func (b *Buffer[T]) Closed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.closed
}

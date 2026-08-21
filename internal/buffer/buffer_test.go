package buffer

import (
	"errors"
	"testing"
)

func TestNewRejectsInvalidCapacity(t *testing.T) {
	_, err := New[int](0)

	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestPushAndPop(t *testing.T) {
	b, err := New[int](2)
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}

	if err := b.Push(10); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	if err := b.Push(20); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	if got := b.Len(); got != 2 {
		t.Fatalf("expected length 2, got %d", got)
	}

	item, ok := b.Pop()
	if !ok {
		t.Fatal("expected item")
	}

	if item != 10 {
		t.Fatalf("expected 10, got %d", item)
	}

	item, ok = b.Pop()
	if !ok {
		t.Fatal("expected item")
	}

	if item != 20 {
		t.Fatalf("expected 20, got %d", item)
	}

	if _, ok := b.Pop(); ok {
		t.Fatal("expected buffer to be empty")
	}
}

func TestPushReturnsErrFull(t *testing.T) {
	b, err := New[int](1)
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}

	if err := b.Push(1); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	err = b.Push(2)

	if !errors.Is(err, ErrFull) {
		t.Fatalf("expected ErrFull, got %v", err)
	}
}

func TestPushAfterClose(t *testing.T) {
	b, err := New[int](1)
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}

	b.Close()

	err = b.Push(1)

	if !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	b, err := New[int](1)
	if err != nil {
		t.Fatalf("failed to create buffer: %v", err)
	}

	b.Close()
	b.Close()

	if !b.Closed() {
		t.Fatal("expected buffer to be closed")
	}
}

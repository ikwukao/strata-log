package pipeline

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		workers    int
		bufferSize int
		handlerNil bool
	}{
		{
			name:       "zero workers",
			workers:    0,
			bufferSize: 1,
		},
		{
			name:       "negative workers",
			workers:    -1,
			bufferSize: 1,
		},
		{
			name:       "zero buffer",
			workers:    1,
			bufferSize: 0,
		},
		{
			name:       "negative buffer",
			workers:    1,
			bufferSize: -1,
		},
		{
			name:       "nil handler",
			workers:    1,
			bufferSize: 1,
			handlerNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler Handler[int]

			if !tt.handlerNil {
				handler = func(context.Context, int) error {
					return nil
				}
			}

			_, err := New(
				ctx,
				tt.workers,
				tt.bufferSize,
				handler,
			)

			if err == nil {
				t.Fatal("expected configuration error")
			}
		})
	}
}

func TestPipelineProcessesItems(t *testing.T) {
	ctx := context.Background()

	var processed atomic.Int32

	p, err := New(
		ctx,
		2,
		10,
		func(context.Context, int) error {
			processed.Add(1)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	for i := 0; i < 10; i++ {
		if err := p.Submit(i); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	p.Close()

	if got := processed.Load(); got != 10 {
		t.Fatalf("expected 10 processed items, got %d", got)
	}
}

func TestPipelineProcessesConcurrently(t *testing.T) {
	ctx := context.Background()

	var active atomic.Int32
	var maxActive atomic.Int32

	p, err := New(
		ctx,
		4,
		10,
		func(context.Context, int) error {
			current := active.Add(1)

			for {
				max := maxActive.Load()

				if current <= max {
					break
				}

				if maxActive.CompareAndSwap(max, current) {
					break
				}
			}

			time.Sleep(20 * time.Millisecond)
			active.Add(-1)

			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	for i := 0; i < 8; i++ {
		if err := p.Submit(i); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	p.Close()

	if got := maxActive.Load(); got < 2 {
		t.Fatalf("expected concurrent processing, max concurrency was %d", got)
	}
}

func TestPipelineReportsHandlerErrors(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("processing failed")

	p, err := New(
		ctx,
		1,
		10,
		func(context.Context, int) error {
			return expectedErr
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	if err := p.Submit(1); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	select {
	case err := <-p.Errors():
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for pipeline error")
	}

	p.Close()
}

func TestPipelineRejectsSubmitAfterClose(t *testing.T) {
	ctx := context.Background()

	p, err := New(
		ctx,
		1,
		1,
		func(context.Context, int) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	p.Close()

	if err := p.Submit(1); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestPipelineRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var processed atomic.Int32

	p, err := New(
		ctx,
		2,
		10,
		func(context.Context, int) error {
			processed.Add(1)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	cancel()

	p.Close()

	if processed.Load() > 10 {
		t.Fatalf("unexpected processing after cancellation")
	}
}

func TestPipelineCloseIsIdempotent(t *testing.T) {
	ctx := context.Background()

	var once sync.Once

	p, err := New(
		ctx,
		1,
		1,
		func(context.Context, int) error {
			once.Do(func() {})
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	p.Close()
	p.Close()
}

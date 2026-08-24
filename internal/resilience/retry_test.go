package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySucceedsAfterFailure(t *testing.T) {
	attempts := 0

	err := Retry(
		context.Background(),
		3,
		time.Millisecond,
		func(context.Context) error {
			attempts++

			if attempts < 3 {
				return errors.New("temporary failure")
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("Retry() error = %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryStopsAfterMaximumAttempts(t *testing.T) {
	attempts := 0

	err := Retry(
		context.Background(),
		3,
		time.Millisecond,
		func(context.Context) error {
			attempts++
			return errors.New("persistent failure")
		},
	)

	if err == nil {
		t.Fatal("expected Retry() to fail")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0

	err := Retry(
		ctx,
		3,
		time.Millisecond,
		func(context.Context) error {
			attempts++
			return errors.New("failure")
		},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if attempts != 0 {
		t.Fatalf("expected zero attempts, got %d", attempts)
	}
}

func TestRetryOperationReceivesContext(t *testing.T) {
	ctx := context.Background()

	err := Retry(
		ctx,
		1,
		time.Millisecond,
		func(operationCtx context.Context) error {
			if operationCtx != ctx {
				t.Fatal("operation received unexpected context")
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
}

func TestRetryRejectsInvalidAttempts(t *testing.T) {
	err := Retry(
		context.Background(),
		0,
		time.Millisecond,
		func(context.Context) error {
			return nil
		},
	)

	if err == nil {
		t.Fatal("expected invalid attempts error")
	}
}

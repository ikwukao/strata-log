// Package resilience provides reliability primitives used by Strata-Log.
package resilience

import (
	"context"
	"errors"
	"time"
)

// Retry executes operation until it succeeds, the attempt limit is reached,
// or the context is canceled.
//
// The first attempt happens immediately. Subsequent attempts use exponential
// backoff starting from the supplied backoff duration.
func Retry(
	ctx context.Context,
	attempts int,
	backoff time.Duration,
	operation func(context.Context) error,
) error {
	if operation == nil {
		return errors.New("retry operation is required")
	}

	if attempts <= 0 {
		return errors.New("retry attempts must be greater than zero")
	}

	var err error
	delay := backoff

	for attempt := 0; attempt < attempts; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}

		err = operation(ctx)
		if err == nil {
			return nil
		}

		if attempt == attempts-1 {
			break
		}

		if delay <= 0 {
			delay = time.Nanosecond
		}

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return ctx.Err()

		case <-timer.C:
		}

		delay *= 2
	}

	return err
}

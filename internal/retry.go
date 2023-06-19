package internal

import (
	"context"
	"log"
	"time"
)

// Effector is a function that performs some action and returns an error.
type Effector func(context.Context, any) error

// Retry is a function that retries effector function for a given number of times with a given delay.
func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context, arg any) error {
		for r := 1; ; r++ {
			err := effector(ctx, arg)
			if err == nil || r > retries {
				return err
			}

			log.Printf("Attempt %d failed, reason %v, retrying in %v", r, err, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

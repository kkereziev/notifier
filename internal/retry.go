package internal

import (
	"context"
	"log"
	"time"
)

type effector func(context.Context, any) error

func retry(effector effector, retries int, delay time.Duration) effector {
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

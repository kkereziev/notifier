package internal

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Effector is a function that performs some action and returns an error.
type Effector func(context.Context, any) error

// Retry is a function that retries effector function for a given number of times with a given delay.
func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context, arg any) error {
		for r := 1; ; r++ {
			err := effector(ctx, arg)
			if err == nil || r >= retries {
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

// RetryGRPC is a function that retries effector function for a given number of times with a given delay.
// It is intended to be used for gRPC calls.
func RetryGRPC(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context, arg any) error {
		for r := 1; ; r++ {
			err := effector(ctx, arg)
			if err == nil || r >= retries || !isGRPCErrorRetriable(err) {
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

// isGRPCErrorRetriable checks if the error is retriable.
func isGRPCErrorRetriable(err error) bool {
	s, _ := status.FromError(err)
	switch s.Code() {
	case codes.NotFound:
		fallthrough
	case codes.FailedPrecondition:
		fallthrough
	case codes.InvalidArgument:
		return false
	default:
		return true
	}
}

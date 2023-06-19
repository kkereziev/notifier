package internal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kkereziev/notifier/internal"
)

func TestRetryPositiveCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name       string
		effector   internal.Effector
		delay      time.Duration
		retries    int
		ctxTimeout time.Duration
	}

	tests := []test{
		{
			name: "function should succeed without retying",
			effector: func(ctx context.Context, a any) error {
				return nil
			},
			delay:      time.Second * 5,
			retries:    2,
			ctxTimeout: time.Second * 10,
		},
		{
			name: "function should succeed within retry and context timeout limit",
			effector: func(ctx context.Context, a any) error {
				incrementor := a.(*int)

				if *incrementor >= 1 {
					return nil
				}

				*incrementor++

				return errors.New("failed")
			},
			delay:      time.Second * 2,
			retries:    2,
			ctxTimeout: time.Second * 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			var counter int

			err := internal.Retry(tc.effector, tc.retries, tc.delay)(ctx, &counter)
			if err != nil {
				t.Errorf("\nExpected error to be nil but got: %s", err)
			}
		})
	}
}

func TestRetryNegativeCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name       string
		effector   internal.Effector
		delay      time.Duration
		retries    int
		ctxTimeout time.Duration
		err        error
	}

	tests := []test{
		{
			name: "timeout of context should stop the execution of effector and return error",
			effector: func(ctx context.Context, a any) error {
				return errors.New("failed")
			},
			delay:      time.Second * 10,
			retries:    3,
			ctxTimeout: time.Second * 1,
			err:        context.DeadlineExceeded,
		},
		{
			name: "function should stop the execution of effector and return error if max retries has been reached",
			effector: func(ctx context.Context, a any) error {
				return errors.New("failed")
			},
			delay:      time.Second * 2,
			retries:    2,
			ctxTimeout: time.Second * 5,
			err:        errors.New("failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			err := internal.Retry(tc.effector, tc.retries, tc.delay)(ctx, "")
			if err.Error() != tc.err.Error() {
				t.Errorf("\nExpected: %s\nActual: %s", tc.err, err)
			}
		})
	}
}

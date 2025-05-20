// Package retry предоставляет функциональность для реализации повторных попыток выполнения операций.
package retry

import (
	"context"
	"time"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sender"
)

type SendFunc func(ctx context.Context, options sender.SendOptions) error

func Retry(sendMetrics SendFunc, retries int) SendFunc {
	return func(ctx context.Context, options sender.SendOptions) error {
		for r := 0; ; r++ {
			nextAttemptAfter := time.Duration(2*r+1) * time.Second
			err := sendMetrics(ctx, options)
			if err == nil || r >= retries {
				return err
			}
			logging.Logg.Info("Attempt %d failed; retrying in %v\n", r+1, nextAttemptAfter)
			select {
			case <-time.After(nextAttemptAfter):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

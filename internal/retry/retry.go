// Package retry предоставляет функциональность для реализации повторных попыток выполнения операций.
package retry

import (
	"context"
	"net/http"
	"time"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/storage"
)

type SendFunc func(ctx context.Context, mStor storage.MemStorage, cl *http.Client, serverAddress string, signKey []byte) error

func Retry(sendMetrics SendFunc, retries int) SendFunc {
	return func(ctx context.Context, mStor storage.MemStorage, cl *http.Client, serverAddress string, signKey []byte) error {
		for r := 0; ; r++ {
			nextAttemptAfter := time.Duration(2*r+1) * time.Second
			err := sendMetrics(ctx, mStor, cl, serverAddress, signKey)
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

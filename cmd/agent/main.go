package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dvkhr/metrix.git/internal/sender"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/dvkhr/metrix.git/internal/storage"
)

type send func(storage.MemStorage, context.Context, *http.Client, string) error

func Retry(sendMetrics send, retries int) send {
	return func(mStor storage.MemStorage, ctx context.Context, cl *http.Client, serverAddress string) error {
		for r := 0; ; r++ {
			nextAttemptAfter := time.Duration(2*r+1) * time.Second
			err := sendMetrics(mStor, ctx, cl, serverAddress)
			if err == nil || r >= retries {
				return err
			}
			fmt.Printf("Attempt %d failed; retrying in %v\n", r+1, nextAttemptAfter)
			select {
			case <-time.After(nextAttemptAfter):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func main() {
	var cfg AgentConfig
	ctx := context.TODO()
	err := cfg.parseFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var mStor storage.MemStorage
	mStor.NewStorage()

	cl := newHTTPClient()

	var collectInterval, sendInterval time.Time
	for {
		if collectInterval.IsZero() ||
			time.Since(collectInterval) >= time.Duration(cfg.pollInterval)*time.Second {
			service.CollectMetrics(ctx, &mStor)
			collectInterval = time.Now()
		}

		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(cfg.reportInterval)*time.Second {
			r := Retry(sender.SendMetrics, 3)
			err = r(mStor, ctx, cl, cfg.serverAddress)
			if err != nil {
				fmt.Println(err)
			}
			sendInterval = time.Now()
		}
		time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
	}
}

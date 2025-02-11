package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func sendMetrics(mStor storage.MemStorage, ctx context.Context, cl *http.Client, cfg Config) error {

	fmt.Printf("+++Send metrics to server+++\n")
	allMetrics, err := mStor.ListSlice(ctx)
	if err == nil {
		jsonMetric, err := json.Marshal(allMetrics)
		if err != nil {
			return err
		}
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		gz.Write(jsonMetric)
		gz.Close()

		req, err := http.NewRequest("POST", buildAllMetricsURL(cfg.serverAddress), &requestBody)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Encoding", "gzip")
		resp, err := cl.Do(req)
		if err == nil {
			fmt.Println(resp.StatusCode)
			defer resp.Body.Close()
			var reader io.ReadCloser
			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					fmt.Println("FAIL create gzip reader: %w", err)
				}
				defer reader.Close()
			default:
				reader = resp.Body
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				fmt.Println("FAIL reader response body: %w", err)
				return err
			}
			fmt.Println(string(body))
		} else {
			return err
		}
		mStor.NewStorage()
	}
	return nil
}

type send func(storage.MemStorage, context.Context, *http.Client, Config) error

func Retry(sendMetrics send, retries int) send {
	return func(mStor storage.MemStorage, ctx context.Context, cl *http.Client, cfg Config) error {
		for r := 0; ; r++ {
			nextAttemptAfter := time.Duration(2*r+1) * time.Second
			err := sendMetrics(mStor, ctx, cl, cfg)
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
	var cfg Config
	ctx := context.TODO()
	err := cfg.parseFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var mStor storage.MemStorage
	mStor.NewStorage()

	cl := &http.Client{Timeout: 5 * time.Second}

	var collectInterval, sendInterval time.Time
	for {
		if collectInterval.IsZero() ||
			time.Since(collectInterval) >= time.Duration(cfg.pollInterval)*time.Second {
			service.CollectMetrics(ctx, &mStor)
			collectInterval = time.Now()
		}

		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(cfg.reportInterval)*time.Second {
			r := Retry(sendMetrics, 3)
			err = r(mStor, ctx, cl, cfg)
			if err != nil {
				fmt.Println(err)
			}
			sendInterval = time.Now()
		}
		time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
	}
}

func buildAllMetricsURL(serverAddress string) string {
	serverURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprint(serverAddress),
		Path:   "updates/",
	}
	return serverURL.String()
}

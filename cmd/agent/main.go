package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func main() {
	serverAddress := flag.String("a", "localhost:8080", "Endpoint HTTP-server")
	reportInterval := flag.Int64("r", 10, "Frequency of sending metrics in seconds")
	pollInterval := flag.Int64("p", 2, "Frequency of metric polling in seconds")
	flag.Parse()
	if *serverAddress == "" {
		*serverAddress = "localhost:8080"

	}
	if *reportInterval == 0 {
		*reportInterval = 10
	}
	if *pollInterval == 0 {
		*pollInterval = 2
	}

	var mStor storage.MemStorage
	mStor.NewMemStorage()

	cl := &http.Client{Timeout: 5 * time.Second}

	var collectInterval, sendInterval time.Time
	for {
		if collectInterval.IsZero() ||
			time.Since(collectInterval) >= time.Duration(*pollInterval)*time.Second {
			metric.CollectMetrics(&mStor)
			collectInterval = time.Now()
		}

		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(*reportInterval)*time.Second {
			fmt.Printf("+++Send metrics to server+++\n")
			metricStrings, err := mStor.MetricStrings()
			if err == nil {
				for _, metricString := range metricStrings {
					err := callURL(cl, buildMetricURL(*serverAddress, metricString))
					if err != nil {
						continue
					}
				}
			}
			sendInterval = time.Now()
		}
		time.Sleep(time.Duration(*pollInterval) * time.Second)
	}
}

func buildMetricURL(serverAddress string, metricString string) string {
	serverURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprint(serverAddress), //"localhost:8080",
		Path:   fmt.Sprintf("update/%s", metricString),
	}

	return serverURL.String()
}

func callURL(cl *http.Client, url string) error {
	res, err := cl.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad http status")
	}

	return nil
}

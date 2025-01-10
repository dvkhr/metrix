package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func main() {

	var mStor storage.MemStorage
	mStor.NewMemStorage()

	iteration := 0

	cl := &http.Client{Timeout: 5 * time.Second}

	for {
		iteration++
		metric.CollectMetrics(&mStor)

		if iteration >= 5 {
			fmt.Printf("+++Send metrics to server+++\n")
			metricStrings, err := mStor.MetricStrings()
			if err == nil {
				for _, metricString := range metricStrings {
					err := callURL(cl, buildMetricURL(metricString))
					if err != nil {
						continue
					}
				}

			}
			iteration = 0
		}
		time.Sleep(2 * time.Second)
	}
}

func buildMetricURL(metricString string) string {
	serverUrl := &url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   fmt.Sprintf("update/%s", metricString),
	}

	return serverUrl.String()
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

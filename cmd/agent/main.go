package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dvkhr/metrix.git/internal/metric"
)

func main() {

	var mStor metric.MemStorage
	mStor.NewMemStorage()

	iteration := 0

	cl := &http.Client{}

	for {
		iteration++
		mStor.CollectMetrics()

		if iteration >= 5 {
			fmt.Printf("+++Send metrics to server+++\n")

			for mtrxName, mtrxVal := range mStor.AllCounterMetrics() {
				err := callURL(cl, buildMetricURL(metric.CounterMetric, mtrxName, fmt.Sprintf("%v", mtrxVal)))
				if err != nil {
					continue
				}
			}

			for mtrxName, mtrxVal := range mStor.AllGaugeMetrics() {
				err := callURL(cl, buildMetricURL(metric.GaugeMetric, mtrxName, fmt.Sprintf("%v", mtrxVal)))
				if err != nil {
					continue
				}
			}
			mStor.ResetCounterMetrics()
			iteration = 0
		}
		time.Sleep(2 * time.Second)
	}
}

func buildMetricURL(mType metric.MetricType, mName string, mValue string) string {
	serverUrl := &url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   fmt.Sprintf("update/%s/%s/%s", string(mType), mName, mValue),
	}

	return serverUrl.String()
}

func callURL(cl *http.Client, url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	res, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad http status")
	}

	return nil
}

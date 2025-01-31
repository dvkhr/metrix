package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func main() {
	var cfg Config
	err := cfg.parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var mStor storage.MemStorage
	mStor.NewMemStorage()

	cl := &http.Client{Timeout: 5 * time.Second}

	var collectInterval, sendInterval time.Time
	for {
		if collectInterval.IsZero() ||
			time.Since(collectInterval) >= time.Duration(cfg.pollInterval)*time.Second {
			metric.CollectMetrics(&mStor)
			collectInterval = time.Now()
		}

		if sendInterval.IsZero() ||
			time.Since(sendInterval) >= time.Duration(cfg.reportInterval)*time.Second {
			fmt.Printf("+++Send metrics to server+++\n")
			allMetrics, err := mStor.List()
			if err == nil {
				for _, metricStruct := range *allMetrics {
					jsonMetric, err := json.Marshal(metricStruct)
					if err != nil {
						continue
					}
					var requestBody bytes.Buffer
					gz := gzip.NewWriter(&requestBody)
					gz.Write(jsonMetric)
					gz.Close()

					req, err := http.NewRequest("POST", buildMetricURL(cfg.serverAddress), &requestBody)
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
							return
						}
						fmt.Println(string(body))
					}

					mStor.NewMemStorage()
				}
			}
			sendInterval = time.Now()
		}
		time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
	}
}

func buildMetricURL(serverAddress string) string {
	serverURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprint(serverAddress),
		Path:   "update/",
	}
	return serverURL.String()
}

func callURL(cl *http.Client, url string, bodyJSON io.Reader) error {

	res, err := cl.Post(url, "application/json", bodyJSON)
	res.Header.Set("Accept-Encoding", "gzip")
	res.Header.Set("Content-Encoding", "gzip")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad http status")
	}
	return nil
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
)

type Config struct {
	serverAddress  string
	reportInterval int64
	pollInterval   int64
}

func (cfg *Config) parseFlags() {
	flag.StringVar(&cfg.serverAddress, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.Int64Var(&cfg.reportInterval, "r", 10, "Frequency of sending metrics in seconds")
	flag.Int64Var(&cfg.pollInterval, "p", 2, "Frequency of metric polling in seconds")
	flag.Parse()
	if envVarAddr := os.Getenv("ADDRESS"); envVarAddr != "" {
		cfg.serverAddress = envVarAddr
	}
	if envVarRep := os.Getenv("REPORT_INTERVAL"); envVarRep != "" {
		cfg.reportInterval, _ = strconv.ParseInt(envVarRep, 10, 64)
	}
	if envVarPoll := os.Getenv("POLL_INTERVAL"); envVarPoll != "" {
		cfg.pollInterval, _ = strconv.ParseInt(envVarPoll, 10, 64)
	}
}
func main() {
	var cfg Config
	cfg.parseFlags()

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
			metricStrings, err := mStor.MetricStrings()
			if err == nil {
				for _, metricString := range metricStrings {
					err := callURL(cl, buildMetricURL(cfg.serverAddress, metricString))
					if err != nil {
						continue
					}
				}
			}
			sendInterval = time.Now()
		}
		time.Sleep(time.Duration(cfg.pollInterval) * time.Second)
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

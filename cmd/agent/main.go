package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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

var ErrIntetrvalNegativ = errors.New("interval is negativ or zero")
var ErrAddressEmpty = errors.New("address is an empty string")

func (cfg *Config) check() error {
	if cfg.serverAddress == "" {
		return ErrAddressEmpty
	} else if cfg.pollInterval <= 0 || cfg.reportInterval <= 0 {
		return ErrIntetrvalNegativ
	} else {
		return nil
	}
}

func (cfg *Config) parseFlags() error {
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
	return cfg.check()
}

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
			allMetrics, err := mStor.AllMetrics()
			if err == nil {
				for _, metricStruct := range *allMetrics {
					jsonMetric, err := json.Marshal(metricStruct)
					if err != nil {
						continue
					}
					err = callURL(cl, buildMetricURL(cfg.serverAddress), bytes.NewBuffer(jsonMetric))
					if err != nil {
						continue
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

func callURL(cl *http.Client, url string, bodyJson io.Reader) error {

	res, err := cl.Post(url, "application/json", bodyJson)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("bad http status")
	}
	return nil
}

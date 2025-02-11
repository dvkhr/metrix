package main

import (
	"errors"
	"flag"
	"os"
	"strconv"
)

type Config struct {
	serverAddress  string
	reportInterval int64
	pollInterval   int64
}

var (
	ErrPollIntetrvalNegativ   = errors.New("poll interval is negativ or zero")
	ErrReportIntetrvalNegativ = errors.New("report interval is negativ or zero")
	ErrAddressEmpty           = errors.New("address is an empty string")
)

func (cfg *Config) check() error {
	var err []error
	if cfg.serverAddress == "" {
		err = append(err, ErrAddressEmpty)
	}
	if cfg.pollInterval <= 0 {
		err = append(err, ErrPollIntetrvalNegativ)
	}
	if cfg.reportInterval <= 0 {
		err = append(err, ErrReportIntetrvalNegativ)
	}

	return errors.Join(err...)
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

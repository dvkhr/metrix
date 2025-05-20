package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type AgentConfig struct {
	serverAddress  string
	reportInterval int64
	pollInterval   int64
	key            string
	rateLimit      int64
	СryptoKey      string
}

var (
	ErrPollIntetrvalNegativ   = errors.New("poll interval is negativ or zero")
	ErrReportIntetrvalNegativ = errors.New("report interval is negativ or zero")
	ErrAddressEmpty           = errors.New("address is an empty string")
	ErrCryptoKeyFileNotFound  = errors.New("crypto key file not found")
)

func (cfg *AgentConfig) check() error {
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

	if cfg.СryptoKey != "" && !fileExists(cfg.СryptoKey) {
		err = append(err, fmt.Errorf("%w: %s", ErrCryptoKeyFileNotFound, cfg.СryptoKey))
	}
	return errors.Join(err...)
}

func (cfg *AgentConfig) parseFlags() error {
	var configFile string

	flag.StringVar(&cfg.serverAddress, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.Int64Var(&cfg.reportInterval, "r", 10, "Frequency of sending metrics in seconds")
	flag.Int64Var(&cfg.pollInterval, "p", 2, "Frequency of metric polling in seconds")
	flag.StringVar(&cfg.key, "k", "", "Key")
	flag.Int64Var(&cfg.rateLimit, "l", 5, "Limiting outgoing requests")
	flag.StringVar(&cfg.СryptoKey, "crypto-key", "", "Path to the public key file for encryption")
	flag.StringVar(&configFile, "c", "", "Path to the JSON configuration file")
	flag.StringVar(&configFile, "config", "", "Path to the JSON configuration file")

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
	if envVarKey := os.Getenv("KEY"); envVarKey != "" {
		cfg.key = envVarKey
	}
	if envVarLim := os.Getenv("RATE_LIMIT"); envVarLim != "" {
		cfg.rateLimit, _ = strconv.ParseInt(envVarLim, 10, 64)
	}
	if envVarCryptoKey := os.Getenv("CRYPTO_KEY"); envVarCryptoKey != "" {
		cfg.СryptoKey = envVarCryptoKey
	}

	if configFileEnv := os.Getenv("CONFIG"); configFileEnv != "" && configFile == "" {
		configFile = configFileEnv
	}

	if configFile != "" {
		if err := cfg.LoadFromFile(configFile); err != nil {
			return fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	return cfg.check()
}
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}
}

type ConfigFile struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func (cfg *AgentConfig) LoadFromFile(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configFile ConfigFile
	if err := json.Unmarshal(file, &configFile); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if configFile.Address != "" && cfg.serverAddress == "" {
		cfg.serverAddress = configFile.Address
	}
	if configFile.ReportInterval != "" && cfg.reportInterval <= 0 {
		duration, err := time.ParseDuration(configFile.ReportInterval)
		if err == nil {
			cfg.reportInterval = int64(duration.Seconds())
		}
	}
	if configFile.PollInterval != "" && cfg.pollInterval <= 0 {
		duration, err := time.ParseDuration(configFile.PollInterval)
		if err == nil {
			cfg.pollInterval = int64(duration.Seconds())
		}
	}
	if configFile.CryptoKey != "" && cfg.СryptoKey == "" {
		cfg.СryptoKey = configFile.CryptoKey
	}

	return nil
}

// fileExists проверяет, существует ли файл по указанному пути.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Package config предоставляет инструменты для загрузки и управления конфигурацией приложения.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ConfigServ struct {
	Address         string
	FileStoragePath string
	StoreInterval   time.Duration
	DBDsn           string
	Restore         bool
	Key             string
	CryptoKey       string
	TrustedSubnet   string
	GRPCAddress     string
}

var (
	ErrStoreIntetrvalNegativ = errors.New("storeInterval is negativ or zero")
	ErrAddressEmpty          = errors.New("address is an empty string")
	ErrCryptoKeyFileNotFound = errors.New("crypto key file not found")
)

func (cfg *ConfigServ) check() error {
	var errs []error
	dirPath := filepath.Dir(cfg.FileStoragePath)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		errs = append(errs, err)
	}
	if len(cfg.Address) == 0 {
		errs = append(errs, ErrAddressEmpty)
	} else if cfg.StoreInterval < 0*time.Microsecond {
		errs = append(errs, ErrStoreIntetrvalNegativ)
	}
	if cfg.CryptoKey != "" {
		if _, err := os.Stat(cfg.CryptoKey); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("%w: %s", ErrCryptoKeyFileNotFound, cfg.CryptoKey))
		}
	}
	return errors.Join(errs...)
}

func (cfg *ConfigServ) ParseFlags() error {
	var storInt int64
	var configFile string

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "The path to the file with metrics")
	flag.StringVar(&cfg.DBDsn, "d", "", "The data source")
	// Next 2 parametrs are useless in current implementation, left here just not to break autotests
	flag.Int64Var(&storInt, "i", 0, "Frequency of saving to disk in seconds")
	flag.BoolVar(&cfg.Restore, "r", true, "loading saved values")
	flag.StringVar(&cfg.Key, "k", "", "Key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to the private key file for decryption (optional)")
	flag.StringVar(&configFile, "c", "", "Path to the JSON configuration file")
	flag.StringVar(&configFile, "config", "", "Path to the JSON configuration file")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet in CIDR format")

	flag.StringVar(&cfg.GRPCAddress, "grpc", "", "Endpoint gRPC-server")

	flag.Parse()

	if envVarAddr := os.Getenv("ADDRESS"); envVarAddr != "" {
		cfg.Address = envVarAddr
	}

	if envVarStor := os.Getenv("FILE_STORAGE_PATH"); envVarStor != "" {
		cfg.FileStoragePath = envVarStor
	}

	if envVarDB := os.Getenv("DATABASE_DSN"); envVarDB != "" {
		cfg.DBDsn = envVarDB
	}
	if envStorInt := os.Getenv("STORE_INTERVAL"); envStorInt != "" {
		storInt, _ = strconv.ParseInt(envStorInt, 10, 64)
	}
	if envReStor := os.Getenv("POLL_INTERVAL"); envReStor != "" {
		cfg.Restore, _ = strconv.ParseBool(envReStor)
	}
	if envVarKey := os.Getenv("KEY"); envVarKey != "" {
		cfg.Key = envVarKey
	}
	cfg.StoreInterval = time.Duration(storInt) * time.Second
	if envVarCryptoKey := os.Getenv("CRYPTO_KEY"); envVarCryptoKey != "" {
		cfg.CryptoKey = envVarCryptoKey
	}
	if envVarTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envVarTrustedSubnet != "" && cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = envVarTrustedSubnet
	}

	if configFileEnv := os.Getenv("CONFIG"); configFileEnv != "" && configFile == "" {
		configFile = configFileEnv
	}

	if configFile != "" {
		if err := cfg.LoadServerConfig(configFile); err != nil {
			return fmt.Errorf("failed to load config from file: %w", err)
		}
	}
	return cfg.check()
}

type LoggerConfig struct {
	LogLevel      string `json:"log_level"`
	ConsoleFormat string `json:"console_format"`
	FileFormat    string `json:"file_format"`
	Destination   string `json:"destination"`
	FilePattern   string `json:"file_pattern"`
}

// LoadLoggerConfig загружает конфигурацию логгера из JSON-файла.
func LoadLoggerConfig(filePath string) (*LoggerConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg LoggerConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// LoadServerConfig загружает конфигурацию сервера из JSON-файла.
func (cfg *ConfigServ) LoadServerConfig(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configFile ServerConfigFile
	if err := json.Unmarshal(file, &configFile); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if configFile.Address != "" && cfg.Address == "localhost:8080" {
		cfg.Address = configFile.Address
	}
	if configFile.Restore && cfg.Restore {
		cfg.Restore = configFile.Restore
	}
	if configFile.StoreInterval != "" && cfg.StoreInterval == 0 {
		duration, err := time.ParseDuration(configFile.StoreInterval)
		if err == nil {
			cfg.StoreInterval = duration
		}
	}
	if configFile.StoreFile != "" && cfg.FileStoragePath == "" {
		cfg.FileStoragePath = configFile.StoreFile
	}
	if configFile.DatabaseDsn != "" && cfg.DBDsn == "" {
		cfg.DBDsn = configFile.DatabaseDsn
	}
	if configFile.CryptoKey != "" && cfg.CryptoKey == "" {
		cfg.CryptoKey = configFile.CryptoKey
	}
	if configFile.TrustedSubnet != "" && cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = configFile.TrustedSubnet
	}

	return nil
}

// Package config предоставляет инструменты для загрузки и управления конфигурацией приложения.
package config

import (
	"encoding/json"
	"errors"
	"flag"
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
}

var (
	ErrStoreIntetrvalNegativ = errors.New("storeInterval is negativ or zero")
	ErrAddressEmpty          = errors.New("address is an empty string")
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
	return errors.Join(errs...)
}

func (cfg *ConfigServ) ParseFlags() error {
	var storInt int64
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "The path to the file with metrics")
	flag.StringVar(&cfg.DBDsn, "d", "", "The data source")
	// Next 2 parametrs are useless in current implementation, left here just not to break autotests
	flag.Int64Var(&storInt, "i", 0, "Frequency of saving to disk in seconds")
	flag.BoolVar(&cfg.Restore, "r", true, "loading saved values")
	flag.StringVar(&cfg.Key, "k", "", "Key")
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

package config

import (
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
}

var ErrStoreIntetrvalNegativ = errors.New("storeInterval is negativ or zero")
var ErrFileStoragePathEmpty = errors.New("fileStoragePath is empty string")
var ErrAddressEmpty = errors.New("address is an empty string")
var ErrNoDirectory = errors.New("no direcrtory in the path")
var ErrDataBaseDsn = errors.New("databasedsn is empty string")

func (cfg *ConfigServ) check() error {
	dirPath := filepath.Dir(cfg.FileStoragePath)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return ErrNoDirectory
	}
	/*if cfg.FileStoragePath == "" {
		return ErrFileStoragePathEmpty
	} else*/if cfg.Address == "" {
		return ErrAddressEmpty
	} else if cfg.StoreInterval < 0*time.Microsecond {
		return ErrStoreIntetrvalNegativ
	} /*else if cfg.DBDsn == "" {
		return ErrDataBaseDsn
	} else*/{
		return nil
	}
}

func (cfg *ConfigServ) ParseFlags() error {
	var storInt int64
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "The path to the file with metrics")
	flag.StringVar(&cfg.DBDsn, "d", "", "The data source")
	// Next 2 parametrs are useless in current implementation, left here just not to break autotests
	flag.Int64Var(&storInt, "i", 0, "Frequency of saving to disk in seconds")
	flag.BoolVar(&cfg.Restore, "r", true, "loading saved values")
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
	cfg.StoreInterval = time.Duration(storInt) * time.Second
	return cfg.check()
}

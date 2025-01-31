package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ConfigServ struct {
	address         string
	fileStoragePath string
	storeInterval   time.Duration
	restore         bool
}

var ErrStoreIntetrvalNegativ = errors.New("storeInterval is negativ or zero")
var ErrFileStoragePathEmpty = errors.New("fileStoragePath is empty string")
var ErrAddressEmpty = errors.New("address is an empty string")
var ErrNoDirectory = errors.New("no direcrtory in the path")

func (cfg *ConfigServ) check() error {
	dirPath := filepath.Dir(cfg.fileStoragePath)
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return ErrNoDirectory
	}
	if cfg.fileStoragePath == "" {
		return ErrFileStoragePathEmpty
	} else if cfg.address == "" {
		return ErrAddressEmpty
	} else if cfg.storeInterval <= 0*time.Microsecond {
		return ErrStoreIntetrvalNegativ
	} else {
		return nil
	}
}

func (cfg *ConfigServ) parseFlags() error {
	var storInt int64
	flag.StringVar(&cfg.address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.fileStoragePath, "f", "metrics.json", "The path to the file with metrics") //"~/go/src/metrix/metrics.json"
	flag.Int64Var(&storInt, "i", 300, "Frequency of saving to disk in seconds")
	flag.BoolVar(&cfg.restore, "r", true, "loading saved values")
	flag.Parse()

	if envVarAddr := os.Getenv("ADDRESS"); envVarAddr != "" {
		cfg.address = envVarAddr
	}

	if envVarStor := os.Getenv("FILE_STORAGE_PATH"); envVarStor != "" {
		cfg.fileStoragePath = envVarStor
	}

	if envStorInt := os.Getenv("STORE_INTERVAL"); envStorInt != "" {
		storInt, _ = strconv.ParseInt(envStorInt, 10, 64)
	}
	if envReStor := os.Getenv("POLL_INTERVAL"); envReStor != "" {
		cfg.restore, _ = strconv.ParseBool(envReStor)
	}
	cfg.storeInterval = time.Duration(storInt) * time.Second
	return cfg.check()
}

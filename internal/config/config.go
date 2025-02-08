package config

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
)

type ConfigServ struct {
	Address         string
	FileStoragePath string
	DBDsn           string
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
	if cfg.Address == "" {
		return ErrAddressEmpty
	} else {
		return nil
	}
}

func (cfg *ConfigServ) ParseFlags() error {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "The path to the file with metrics")
	flag.StringVar(&cfg.DBDsn, "d", "", "The data source")
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

	return cfg.check()
}

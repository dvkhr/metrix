package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logger"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

type ConfigServ struct {
	address         string
	fileStoragePath string
	storeInterval   int64
	restore         bool
}

var ErrStoreIntetrvalNegativ = errors.New("storeInterval is negativ or zero")
var ErrFileStoragePathEmpty = errors.New("fileStoragePath is empty string")
var ErrAddressEmpty = errors.New("address is an empty string")

func (cfg *ConfigServ) check() error {
	if cfg.fileStoragePath == "" {
		return ErrFileStoragePathEmpty
	} else if cfg.address == "" {
		return ErrAddressEmpty
	} else if cfg.storeInterval <= 0 {
		return ErrStoreIntetrvalNegativ
	} else {
		return nil
	}
}

func (cfg *ConfigServ) parseFlags() error {
	flag.StringVar(&cfg.address, "a", "localhost:8080", "Endpoint HTTP-server")
	flag.StringVar(&cfg.fileStoragePath, "f", "metrics.json", "The path to the file with metrics") //"~/go/src/metrix/metrics.json"
	flag.Int64Var(&cfg.storeInterval, "i", 300, "Frequency of saving to disk in seconds")
	flag.BoolVar(&cfg.restore, "r", true, "loading saved values")
	flag.Parse()

	if envVarAddr := os.Getenv("ADDRESS"); envVarAddr != "" {
		cfg.address = envVarAddr
	}

	if envVarStor := os.Getenv("FILE_STORAGE_PATH"); envVarStor != "" {
		cfg.fileStoragePath = envVarStor
	}

	if envStorInt := os.Getenv("STORE_INTERVAL"); envStorInt != "" {
		cfg.storeInterval, _ = strconv.ParseInt(envStorInt, 10, 64)
	}
	if envReStor := os.Getenv("POLL_INTERVAL"); envReStor != "" {
		cfg.restore, _ = strconv.ParseBool(envReStor)
	}
	return cfg.check()
}

func main() {
	var cfg ConfigServ
	err := cfg.parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	MetricServer := handlers.NewMetricsServer(&storage.MemStorage{})

	if cfg.restore {
		file, err := os.OpenFile(cfg.fileStoragePath, os.O_RDONLY, 0666)
		if err != nil {
			logger.Sugar.Errorln("unable open file", "file", cfg.fileStoragePath, "error", err)
		} else {
			err = handlers.RestoreMetrics(MetricServer, file)
			if err != nil {
				logger.Sugar.Errorln("unable to restore metrics", "error", err)
			} else {
				logger.Sugar.Infoln("metrics restored")
			}
			file.Close()
			if err != nil {
				logger.Sugar.Errorln("unable close file", "error", err)
			}
		}
	}

	r := chi.NewRouter()

	r.Get("/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandleGetAllMetrics)))
	r.Get("/value/{type}/{name}", logger.WithLogging(MetricServer.HandleGetMetric))
	r.Post("/value/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandleGetMetricJSON)))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandlePutMetricJSON)))
		r.Post("/*", logger.WithLogging(MetricServer.IncorrectMetricRq))
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", logger.WithLogging(MetricServer.NotfoundMetricRq))
			r.Post("/{name}/{value}", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandlePutGaugeMetric)))
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", logger.WithLogging(MetricServer.NotfoundMetricRq))
			r.Post("/{name}/{value}", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandlePutCounterMetric)))
		})
	})

	server := &http.Server{Addr: cfg.address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	dumpMetrics := func() {
		file, err := os.OpenFile(cfg.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			logger.Sugar.Errorln("unable open file", "file", cfg.fileStoragePath, "error", err)
		}

		err = handlers.DumpMetrics(MetricServer, file)
		if err != nil {
			logger.Sugar.Errorln("unable dump metrics", "error", err)
		}
		err = file.Sync()
		if err != nil {
			logger.Sugar.Errorln("unable sync file", "error", err)
		}
		err = file.Close()
		if err != nil {
			logger.Sugar.Errorln("unable close file", "error", err)
		}

		logger.Sugar.Infoln("metrics dumped")
	}

	go func() {
		for {
			time.Sleep(time.Duration(cfg.storeInterval) * time.Second)
			dumpMetrics()
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar.Fatalw(err.Error(), "event", "start server")
		}
	}()

	<-stop
	logger.Sugar.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Sugar.Fatalw(err.Error(), "event", "server forced to shutdown", "error", err)
	}

	dumpMetrics()

	logger.Sugar.Infoln("server shut down completed")
}

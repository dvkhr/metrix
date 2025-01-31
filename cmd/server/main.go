package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logger"
	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	var cfg ConfigServ
	err := cfg.parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var mutex sync.Mutex

	MetricServer := handlers.NewMetricsServer(&storage.MemStorage{})
	if cfg.storeInterval == 0*time.Second {
		MetricServer.Sync = true
	}

	if cfg.restore {
		file, err := os.OpenFile(cfg.fileStoragePath, os.O_RDONLY, 0666)
		if err != nil {
			logger.Sugar.Errorln("unable open file", "file", cfg.fileStoragePath, "error", err)
		} else {
			err = metric.RestoreMetrics(MetricServer.MetricStorage, file)
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
	r.Post("/value/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.ExtractMetric)))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.UpdateMetric)))
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
		mutex.Lock()
		err = metric.DumpMetrics(MetricServer.MetricStorage, file)
		if err != nil {
			logger.Sugar.Errorln("unable dump metrics", "error", err)
		}
		err = file.Sync()
		if err != nil {
			logger.Sugar.Errorln("unable sync file", "error", err)
		}
		mutex.Unlock()
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

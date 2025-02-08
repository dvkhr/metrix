package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	var cfg config.ConfigServ
	err := cfg.ParseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	MetricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		logger.Sugar.Errorln("unable to initialize storage", "error", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Get("/", logger.WithLogging(gzip.GzipMiddleware(MetricServer.HandleGetAllMetrics)))
	r.Get("/value/{type}/{name}", logger.WithLogging(MetricServer.HandleGetMetric))
	r.Get("/ping", logger.WithLogging(MetricServer.CheckDBConnect))
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

	server := &http.Server{Addr: cfg.Address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

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

	MetricServer.MetricStorage.FreeStorage()

	logger.Sugar.Infoln("server shut down completed")
}

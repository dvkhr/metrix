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
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sign"
	"github.com/go-chi/chi/v5"
)

func main() {
	logging.Logg = logging.NewLogger("debug", "text", "json", "both", "logs/2006-01-02.log")
	if logging.Logg == nil {
		fmt.Println("Failed to initialize logger")
		os.Exit(1)
	}

	var cfg config.ConfigServ
	err := cfg.ParseFlags()
	if err != nil {
		logging.Logg.Error("Server configuration error: %v", err)
		os.Exit(1)
	}

	MetricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		logging.Logg.Error("unable to initialize storage: %v", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(logging.LoggingMiddleware(logging.Logg))

	r.Get("/", gzip.GzipMiddleware(MetricServer.HandleGetAllMetrics))
	r.Get("/value/{type}/{name}", MetricServer.HandleGetMetric)
	r.Get("/ping", MetricServer.CheckDBConnect)
	r.Post("/value/", gzip.GzipMiddleware(MetricServer.ExtractMetric))
	r.Post("/updates/", gzip.GzipMiddleware(sign.SignCheck(MetricServer.UpdateBatch, []byte(MetricServer.Config.Key))))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", gzip.GzipMiddleware(MetricServer.UpdateMetric))
		r.Post("/*", MetricServer.IncorrectMetricRq)
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", MetricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", gzip.GzipMiddleware(MetricServer.HandlePutGaugeMetric))
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", MetricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", gzip.GzipMiddleware(MetricServer.HandlePutCounterMetric))
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
			logging.Logg.Error("Server failed to start", "error", err)
		}
	}()

	<-stop
	logging.Logg.Info("Shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logging.Logg.Error("Server shutdown error", "error", err)
	}

	MetricServer.MetricStorage.FreeStorage()

	logging.Logg.Info("Server stopped")
}

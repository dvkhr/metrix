package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/routes"

	_ "net/http/pprof" // Импортируем pprof
)

var (
	cfg          config.ConfigServ
	MetricServer *handlers.MetricsServer
	server       *http.Server
)

func init() {
	handlers.CheckImplementations()

	logging.Logg = logging.NewLogger("debug", "text", "json", "both", "logs/2006-01-02.log")
	if logging.Logg == nil {
		fmt.Println("Failed to initialize logger")
		os.Exit(1)
	}

	go startPProf()

	if err := cfg.ParseFlags(); err != nil {
		logging.Logg.Error("Failed to parse configuration: %v", err)
		os.Exit(1)
	}

	var err error
	MetricServer, err = handlers.NewMetricsServer(cfg)
	if err != nil {
		logging.Logg.Error("Unable to initialize storage: %v", err)
		os.Exit(1)
	}

	r := routes.SetupRoutes(cfg, MetricServer)

	server = &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func main() {
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

func startPProf() {
	fmt.Println("Starting pprof server on :9090")
	if err := http.ListenAndServe("localhost:9090", nil); err != nil {
		fmt.Println("Failed to start pprof server:", err)
	}
}

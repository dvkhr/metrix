package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/dvkhr/metrix.git/internal/buildinfo"
	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/routes"
	"github.com/go-chi/chi/v5"

	_ "net/http/pprof" // Импортируем pprof
)

var buildVersion string
var buildDate string
var buildCommit string

var (
	cfg          config.ConfigServ
	MetricServer *handlers.MetricsServer
	server       *http.Server
)

func init() {
	handlers.CheckImplementations()

	var err error

	// Установка рабочей директории в корень проекта
	exeDir := filepath.Dir(os.Args[0])             // Директория исполняемого файла
	projectRoot := filepath.Join(exeDir, "../../") // Поднимаемся на два уровня выше
	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("Failed to change working directory to %s: %v", projectRoot, err)
		os.Exit(1)
	}
	// Инициализация глобального логгера
	if err = logging.InitLogger("internal/config/logger_config.json"); err != nil {
		// Логируем ошибку и завершаем программу с кодом ошибки
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1) // Завершаем программу с кодом ошибки
	}
	go startPProf()

	if err = cfg.ParseFlags(); err != nil {
		logging.Logg.Error("Failed to parse configuration: %v", err)
		os.Exit(1)
	}

	MetricServer, err = handlers.NewMetricsServer(cfg)
	if err != nil {
		logging.Logg.Error("Unable to initialize storage: %v", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r = routes.SetupRoutes(r, logging.Logg, cfg, MetricServer)

	server = &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func main() {
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

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

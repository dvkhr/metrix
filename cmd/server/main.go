package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/grpcserver"
	"github.com/dvkhr/metrix.git/internal/service"

	"github.com/dvkhr/metrix.git/internal/buildinfo"
	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/routes"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "net/http/pprof" // Импортируем pprof
)

var buildVersion string
var buildDate string
var buildCommit string

var (
	cfg          config.ConfigServ
	MetricServer *handlers.MetricsServer
	server       *http.Server
	grpcServer   *grpc.Server
)

// ./server -crypto-key= "/home/max/go/src/metrix/cmd/server/private_key.pem"
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

	logConfiguration(cfg)
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

func logConfiguration(cfg config.ConfigServ) {
	logging.Logg.Info("Configuration loaded:\n" +
		"  Address: " + cfg.Address + "\n" +
		"  FileStoragePath: " + cfg.FileStoragePath + "\n" +
		"  StoreInterval: " + cfg.StoreInterval.String() + "\n" +
		"  DBDsn: " + cfg.DBDsn + "\n" +
		"  Restore: " + fmt.Sprintf("%v", cfg.Restore) + "\n" +
		"  Key: " + logging.MaskSensitiveData(cfg.Key) + "\n" +
		"  CryptoKey: " + logging.MaskSensitiveData(cfg.CryptoKey) + "\n" +
		"  TrustedSubnet: " + cfg.TrustedSubnet + "\n" +
		"  GRPCAddress: " + cfg.GRPCAddress,
	)
}

func main() {
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logg.Error("Server failed to start", "error", err)
		}
	}()

	if cfg.GRPCAddress != "" {
		go startGRPCServer(cfg.GRPCAddress, MetricServer.MetricStorage, []byte(cfg.Key))
	}
	<-stop
	logging.Logg.Info("Shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logging.Logg.Error("Server shutdown error", "error", err)
	}
	if grpcServer != nil {
		grpcServer.GracefulStop()
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

func startGRPCServer(address string, metricStorage service.MetricStorage, signKey []byte) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logging.Logg.Error("Failed to listen on gRPC address: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Регистрация сервиса MetricsService
	pb.RegisterMetricsServiceServer(grpcServer, &grpcserver.MetricsServer{
		MetricStorage: metricStorage,
		SignKey:       signKey,
	})

	// Включение Server Reflection
	reflection.Register(grpcServer)

	logging.Logg.Info("Starting gRPC server with reflection on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		logging.Logg.Error("Failed to start gRPC server: %v", err)
	}
}

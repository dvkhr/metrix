package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/dvkhr/metrix.git/internal/buildinfo"
	"github.com/dvkhr/metrix.git/internal/crypto"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sender"
	"github.com/dvkhr/metrix.git/internal/service"
)

var buildVersion string
var buildDate string
var buildCommit string

// go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=2025-05-05 -X main.buildCommit=commit"
//
//	./agent -crypto-key= "/home/max/go/src/metrix/cmd/agent/public_key.pem"
func main() {
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	// Установка рабочей директории в корень проекта
	exeDir := filepath.Dir(os.Args[0])             // Директория исполняемого файла
	projectRoot := filepath.Join(exeDir, "../../") // Поднимаемся на два уровня выше
	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("Failed to change working directory to %s: %v", projectRoot, err)
		return
	}

	// Инициализация глобального логгера
	if err := logging.InitLogger("internal/config/logger_config.json"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	var cfg AgentConfig
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := cfg.parseFlags()

	if err != nil {
		logging.Logg.Error("Server configuration error: %v", err)
		return
	}

	// Чтение публичного ключа
	var publicKey *rsa.PublicKey
	if cfg.СryptoKey != "" {
		var err error
		publicKey, err = crypto.ReadPublicKey(cfg.СryptoKey)
		if err != nil {
			logging.Logg.Error("Failed to read public key: %v", err)
			return
		}
		logging.Logg.Info("Public key successfully loaded")
	}

	cl := newHTTPClient()

	stopChan := make(chan bool)
	defer close(stopChan)

	payloadChan := make(chan service.Metrics)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	collectOSWorker := CollectWorker{wf: service.CollectMetricsOS, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	collectChWorker := CollectWorker{wf: service.CollectMetricsCh, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	sendMetricsWorker := SendWorker{wf: sender.SendMetrics, poll: cfg.reportInterval, ctx: ctx, payloadChan: payloadChan,
		stopChan: stopChan, cl: cl, serverAddress: cfg.serverAddress, signKey: []byte(cfg.key), publicKey: publicKey}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		collectOSWorker.StartCollecting()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		collectChWorker.StartCollecting()
	}()

	for i := 0; i < int(cfg.rateLimit); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetricsWorker.Run()
		}()
	}

	<-signalChan
	logging.Logg.Info("shutting down agent...")

	cancel() // Отменяем контекст
	stopChan <- true

	wg.Wait() // Ожидаем завершения всех рабочих процессов

	close(payloadChan)
	for metric := range payloadChan {
		logging.Logg.Warn("Unsent metric detected", "metric", metric)
	}

	logging.Logg.Info("agent shut down completed")
}

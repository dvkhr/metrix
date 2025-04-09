package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sender"
	"github.com/dvkhr/metrix.git/internal/service"
)

func main() {
	logging.Logg = logging.NewLogger("debug", "text", "json", "both", "logs/2006-01-02.log")
	if logging.Logg == nil {
		fmt.Println("Failed to initialize logger")
		os.Exit(1)
	}
	var cfg AgentConfig
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := cfg.parseFlags()

	if err != nil {
		logging.Logg.Error("Server configuration error: %v", err)
		os.Exit(1)
	}

	cl := newHTTPClient()

	stopChan := make(chan bool)
	defer close(stopChan)

	payloadChan := make(chan service.Metrics)
	defer close(payloadChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	collectOSWorker := CollectWorker{wf: service.CollectMetricsOS, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	collectChWorker := CollectWorker{wf: service.CollectMetricsCh, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	sendMetricsWorker := SendWorker{wf: sender.SendMetrics, poll: cfg.reportInterval, ctx: ctx,
		payloadChan: payloadChan, stopChan: stopChan, cl: cl, serverAddress: cfg.serverAddress, signKey: []byte(cfg.key)}

	go collectOSWorker.StartCollecting()
	go collectChWorker.StartCollecting()

	for i := 0; i < int(cfg.rateLimit); i++ {
		go sendMetricsWorker.Run()
	}

	<-signalChan
	logging.Logg.Info("shutting down agent...")

	stopChan <- true

	logging.Logg.Info("agent shut down completed")
}

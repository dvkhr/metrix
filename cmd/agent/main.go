package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dvkhr/metrix.git/internal/sender"
	"github.com/dvkhr/metrix.git/internal/service"
)

func main() {
	var cfg AgentConfig
	ctx := context.TODO()
	err := cfg.parseFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cl := newHTTPClient()

	stopChan := make(chan bool)
	defer close(stopChan)

	payloadChan := make(chan service.Metrics)
	defer close(payloadChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	CollectOSWorker := CollectWorker{wf: service.CollectMetricsOS, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	CollectChWorker := CollectWorker{wf: service.CollectMetricsCh, poll: cfg.pollInterval, ctx: ctx, payloadChan: payloadChan, stopChan: stopChan}
	SendMetricsWorker := SendWorker{wf: sender.SendMetrics, poll: cfg.reportInterval, ctx: ctx,
		payloadChan: payloadChan, stopChan: stopChan, cl: cl, serverAddress: cfg.serverAddress, signKey: []byte(cfg.key)}

	go CollectOSWorker.Run()
	go CollectChWorker.Run()

	for i := 0; i < int(cfg.rateLimit); i++ {
		go SendMetricsWorker.Run()
	}

	<-signalChan
	fmt.Println("shutting down agent...")

	stopChan <- true

	fmt.Println("agent shut down completed")
}

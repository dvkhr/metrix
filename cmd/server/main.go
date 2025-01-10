package main

import (
	"net/http"

	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func main() {
	MetricServer := handlers.NewMetricsServer(&storage.MemStorage{})

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/{name}/{value}", MetricServer.HandlePutGaugeMetric)
	mux.HandleFunc("/update/counter/{name}/{value}", MetricServer.HandlePutCounterMetric)
	mux.HandleFunc("/update/gauge/", MetricServer.NotfoundMetricRq)
	mux.HandleFunc("/update/counter/", MetricServer.NotfoundMetricRq)
	mux.HandleFunc("/update/", MetricServer.IncorrectMetricRq)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	var netAddress *string
	envVarAddr, addrB := os.LookupEnv("ADDRESS")
	if addrB {
		netAddress = &envVarAddr
	} else {
		netAddress = flag.String("a", "localhost:8080", "Endpoint HTTP-server")
		flag.Parse()
	}
	MetricServer := handlers.NewMetricsServer(&storage.MemStorage{})
	r := chi.NewRouter()
	r.Get("/", MetricServer.HandleGetAllMetrics)
	r.Get("/value/{type}/{name}", MetricServer.HandleGetMetric)
	r.Route("/update", func(r chi.Router) {
		r.Post("/*", MetricServer.IncorrectMetricRq)
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", MetricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", MetricServer.HandlePutGaugeMetric)
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", MetricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", MetricServer.HandlePutCounterMetric)
		})
	})
	err := http.ListenAndServe(*netAddress, r)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"net/http"

	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
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
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		panic(err)
	}
}

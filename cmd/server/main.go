package main

import (
	"flag"
	"net/http"

	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logger"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	netAddress := flag.String("a", "localhost:8080", "Endpoint HTTP-server")
	flag.Parse()
	MetricServer := handlers.NewMetricsServer(&storage.MemStorage{})
	r := chi.NewRouter()
	r.Get("/", logger.WithLogging(MetricServer.HandleGetAllMetrics))
	r.Get("/value/{type}/{name}", logger.WithLogging(MetricServer.HandleGetMetric))
	r.Route("/update", func(r chi.Router) {
		r.Post("/*", logger.WithLogging(MetricServer.IncorrectMetricRq))
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", logger.WithLogging(MetricServer.NotfoundMetricRq))
			r.Post("/{name}/{value}", logger.WithLogging(MetricServer.HandlePutGaugeMetric))
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", logger.WithLogging(MetricServer.NotfoundMetricRq))
			r.Post("/{name}/{value}", logger.WithLogging(MetricServer.HandlePutCounterMetric))
		})
	})
	if err := http.ListenAndServe(*netAddress, r); err != nil {
		logger.Sugar.Fatalw(err.Error(), "event", "start server")
	}
}

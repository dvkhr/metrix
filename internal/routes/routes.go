package routes

import (
	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sign"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(cfg config.ConfigServ, metricServer *handlers.MetricsServer) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(logging.LoggingMiddleware(logging.Logg))

	// Routes
	r.Get("/", gzip.GzipMiddleware(metricServer.HandleGetAllMetrics))
	r.Get("/value/{type}/{name}", metricServer.HandleGetMetric)
	r.Get("/ping", metricServer.CheckDBConnect)
	r.Post("/value/", gzip.GzipMiddleware(metricServer.ExtractMetric))
	r.Post("/updates/", gzip.GzipMiddleware(sign.SignCheck(metricServer.UpdateBatch, []byte(metricServer.Config.Key))))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", gzip.GzipMiddleware(metricServer.UpdateMetric))
		r.Post("/*", metricServer.IncorrectMetricRq)
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", metricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", gzip.GzipMiddleware(metricServer.HandlePutGaugeMetric))
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", metricServer.NotfoundMetricRq)
			r.Post("/{name}/{value}", gzip.GzipMiddleware(metricServer.HandlePutCounterMetric))
		})
	})

	return r
}

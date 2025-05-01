// Package routes предоставляет инструменты для настройки маршрутов HTTP-сервера метрик.
package routes

import (
	"net/http"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sign"
	"github.com/go-chi/chi/v5"
)

// MetricServer определяет методы, которые должны быть реализованы сервером метрик.
type MetricServer interface {
	HandleGetAllMetrics(w http.ResponseWriter, r *http.Request)
	HandleGetMetric(w http.ResponseWriter, r *http.Request)
	CheckDBConnect(w http.ResponseWriter, r *http.Request)
	ExtractMetric(w http.ResponseWriter, r *http.Request)
	UpdateBatch(w http.ResponseWriter, r *http.Request)
	UpdateMetric(w http.ResponseWriter, r *http.Request)
	IncorrectMetricRq(w http.ResponseWriter, r *http.Request)
	NotfoundMetricRq(w http.ResponseWriter, r *http.Request)
	HandlePutGaugeMetric(w http.ResponseWriter, r *http.Request)
	HandlePutCounterMetric(w http.ResponseWriter, r *http.Request)
}

// SetupRoutes настраивает маршруты HTTP-сервера для обработки запросов метрик.
//
// Параметры:
// - r: маршрутизатор chi
// - logger: логгер
// - cfg: Конфигурация сервера, содержащая параметры для настройки маршрутов.
// - metricServer: Экземпляр MetricsServer, который обрабатывает логику работы с метриками.
//
// Логика работы:
// 1. Создается новый маршрутизатор chi.
// 2. Добавляется middleware для логирования всех запросов.
// 3. Настраиваются маршруты:
//   - GET "/": Возвращает HTML-страницу со всеми метриками.
//   - GET "/value/{type}/{name}": Получает значение метрики по её типу и имени.
//   - GET "/ping": Проверяет подключение к базе данных.
//   - POST "/value/": Извлекает метрику из JSON-тела запроса и помещает ее в хранилище.
//   - POST "/updates/": Обновляет метрики пакетно с возможностью проверки подписи.
//   - POST "/update/*": Обрабатывает некорректные запросы на обновление метрик.
//   - POST "/update/gauge/{name}/{value}": Обновляет метрику типа "gauge".
//   - POST "/update/counter/{name}/{value}": Обновляет метрику типа "counter".
//
// 4. Для некоторых маршрутов применяется middleware GzipMiddleware для сжатия ответов.
// 5. Для маршрута "/updates/" также применяется middleware SignCheck для проверки подписи запроса.
//
// Возвращаемое значение:
// - *chi.Mux: Настроенный маршрутизатор chi с определенными маршрутами.
func SetupRoutes(r *chi.Mux, logger *logging.Logger, cfg config.ConfigServ, metricServer MetricServer) *chi.Mux {

	// Middleware
	r.Use(logging.LoggingMiddleware(logging.Logg))

	// Routes
	r.Get("/", gzip.GzipMiddleware(metricServer.HandleGetAllMetrics))
	r.Get("/value/{type}/{name}", metricServer.HandleGetMetric)
	r.Get("/ping", metricServer.CheckDBConnect)
	r.Post("/value/", gzip.GzipMiddleware(metricServer.ExtractMetric))
	r.Post("/updates/", gzip.GzipMiddleware(sign.SignCheck(metricServer.UpdateBatch, []byte(cfg.Key))))
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

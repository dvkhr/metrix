package routes

import (
	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/sign"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes настраивает маршруты HTTP-сервера для обработки запросов метрик.
//
// Функция использует библиотеку chi для создания маршрутизатора и определяет
// набор endpoints для работы с метриками (получение, обновление, проверка подключения и т.д.).
//
// Параметры:
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

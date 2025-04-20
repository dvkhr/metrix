package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"text/template"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/dvkhr/metrix.git/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MetricStorage представляет интерфейс для работы с хранилищем метрик.
//
// Он определяет набор методов для сохранения, получения и управления метриками.
// Реализации этого интерфейса могут использовать различные типы хранилищ,
// такие как база данных, файловое хранилище или оперативная память.
//
// Методы:
// - Save: Сохраняет одну метрику в хранилище.
// - SaveAll: Сохраняет массив метрик в хранилище.
// - Get: Получает метрику по её имени.
// - List: Возвращает все метрики в виде мапы, где ключ — имя метрики.
// - ListSlice: Возвращает все метрики в виде слайса.
// - NewStorage: Инициализирует хранилище.
// - FreeStorage: Освобождает ресурсы, связанные с хранилищем.
// - CheckStorage: Проверяет доступность хранилища.
type MetricStorage interface {
	Save(ctx context.Context, mt service.Metrics) error
	SaveAll(ctx context.Context, mt *[]service.Metrics) error
	Get(ctx context.Context, metricName string) (*service.Metrics, error)
	List(ctx context.Context) (*map[string]service.Metrics, error)
	ListSlice(ctx context.Context) ([]service.Metrics, error)
	NewStorage() error
	FreeStorage() error
	CheckStorage() error
}

//mockgen -source=internal/handlers/handlers.go -destination=internal/mocks/mock_storage.go -package=mocks

// MetricsServer представляет сервер для обработки метрик.
// Он управляет хранилищем метрик и предоставляет методы для их сохранения, получения и обработки.
// Поля:
//   - MetricStorage: Интерфейс хранилища метрик (база данных, файловое хранилище или память).
//     Используется для выполнения операций с метриками.
//   - Config: Конфигурация сервера, содержащая параметры подключения и настройки.
//   - syncMutex: Мьютекс для обеспечения потокобезопасности при работе с общими ресурсами.
type MetricsServer struct {
	MetricStorage MetricStorage
	Config        config.ConfigServ
	syncMutex     sync.Mutex
}

// NewMetricsServer создает новый экземпляр MetricsServer с выбранным хранилищем метрик.
//
// Выбор хранилища зависит от конфигурации:
// - Если Config.DBDsn не пустой, используется хранилище на основе PostgreSQL (DBStorage).
// - Если Config.FileStoragePath не пустой, используется файловое хранилище (FileStorage).
// - Если ни один из вышеперечисленных параметров не задан, используется хранилище в оперативной памяти (MemStorage).
//
// После выбора хранилища вызывается метод NewStorage для его инициализации.
// Если инициализация завершается ошибкой, она возвращается вызывающей стороне.
//
// Параметры:
// - Config: Конфигурация сервера, содержащая параметры для подключения к хранилищу.
//
// Возвращаемые значения:
// - *MetricsServer: Указатель на созданный экземпляр MetricsServer.
// - error: Ошибка, если произошла проблема при инициализации хранилища.
func NewMetricsServer(Config config.ConfigServ) (*MetricsServer, error) {
	var ms MetricStorage
	if len(Config.DBDsn) > 0 {
		ms = &storage.DBStorage{DBDSN: Config.DBDsn}

	} else if len(Config.FileStoragePath) > 0 {
		ms = &storage.FileStorage{FileStoragePath: Config.FileStoragePath}

	} else {
		ms = &storage.MemStorage{}
	}

	if err := ms.NewStorage(); err != nil {
		return nil, err
	}

	return &MetricsServer{MetricStorage: ms, Config: Config}, nil
}

// IncorrectMetricRq обрабатывает некорректные запросы на обновление метрик.
//
// В ответ на запрос отправляется HTTP-ошибка с кодом 400 (Bad Request) и сообщением:
// "Incorrect update metric request!".
func (ms *MetricsServer) IncorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

// NotfoundMetricRq обрабатывает запросы на получение или обновление несуществующих метрик.
//
// В ответ на запрос отправляется HTTP-ошибка с кодом 404 (Not Found) и сообщением:
// "Metric not found!".
func (ms *MetricsServer) NotfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
}

// HandlePutGaugeMetric обрабатывает HTTP-запросы на сохранение метрики типа "gauge".
//
// Метод извлекает имя метрики и её значение из параметров запроса, проверяет их корректность,
// и сохраняет метрику в хранилище.
//
// Логика работы:
//  1. Используется мьютекс для обеспечения потокобезопасности при работе с хранилищем.
//  2. Извлекается имя метрики из параметра пути "name". Если имя пустое, возвращается ошибка 404 (Not Found).
//  3. Извлекается значение метрики из параметра пути "value" и преобразуется в float64.
//     Если значение некорректно, возвращается ошибка 400 (Bad Request).
//  4. Создается объект Metrics с типом "gauge" и сохраняется в хранилище.
//  5. После сохранения выполняется запрос метрики из хранилища для подтверждения успешности операции.
//  6. В случае успеха возвращается HTTP-статус 200 (OK).
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий параметры пути "name" и "value".
func (ms *MetricsServer) HandlePutGaugeMetric(res http.ResponseWriter, req *http.Request) {
	ms.syncMutex.Lock()
	defer ms.syncMutex.Unlock()

	ctx := context.TODO()

	n := req.PathValue("name")
	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}
	v, err := strconv.ParseFloat(req.PathValue("value"), 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	mTemp := &service.Metrics{}
	mTemp.ID = n

	vtemp := service.GaugeMetricValue(v)
	mTemp.Value = &vtemp
	mTemp.MType = service.GaugeMetric

	ms.MetricStorage.Save(ctx, *mTemp)
	ms.MetricStorage.Get(ctx, req.PathValue("name"))
	res.WriteHeader(http.StatusOK)
}

// HandlePutCounterMetric обрабатывает HTTP-запросы на сохранение метрики типа "counter".
//
// Метод извлекает имя метрики и её значение из параметров запроса, проверяет их корректность,
// и сохраняет метрику в хранилище.
//
// Логика работы:
//  1. Используется мьютекс для обеспечения потокобезопасности при работе с хранилищем.
//  2. Извлекается имя метрики из параметра пути "name". Если имя пустое, возвращается ошибка 404 (Not Found).
//  3. Извлекается значение метрики из параметра пути "value" и преобразуется в int64.
//     Если значение некорректно, возвращается ошибка 400 (Bad Request).
//  4. Создается объект Metrics с типом "counter" и сохраняется в хранилище.
//  5. После сохранения выполняется запрос метрики из хранилища для подтверждения успешности операции.
//  6. В случае успеха возвращается HTTP-статус 200 (OK).
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий параметры пути "name" и "value".
func (ms *MetricsServer) HandlePutCounterMetric(res http.ResponseWriter, req *http.Request) {
	ms.syncMutex.Lock()
	defer ms.syncMutex.Unlock()

	ctx := context.TODO()

	n := req.PathValue("name")
	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}
	v, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	mTemp := &service.Metrics{}
	mTemp.ID = n

	vtemp := service.CounterMetricValue(v)
	mTemp.Delta = &vtemp
	mTemp.MType = service.CounterMetric

	ms.MetricStorage.Save(ctx, *mTemp)
	ms.MetricStorage.Get(ctx, req.PathValue("name"))
	res.WriteHeader(http.StatusOK)
}

// UpdateMetric обрабатывает HTTP-запросы на обновление метрик через JSON.
//
// Метод принимает метрику в формате JSON, проверяет её корректность,
// сохраняет в хранилище и возвращает обновленную метрику в ответе.
//
// Логика работы:
//  1. Используется мьютекс для обеспечения потокобезопасности при работе с хранилищем.
//  2. Проверяется, что метод запроса является POST. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  3. Тело запроса считывается и десериализуется в объект Metrics.
//     Если данные некорректны, возвращается ошибка 400 (Bad Request).
//  4. Метрика сохраняется в хранилище. Если сохранение завершается ошибкой, возвращается ошибка 400 (Bad Request).
//  5. После сохранения метрика запрашивается из хранилища для подтверждения успешности операции.
//  6. Обновленная метрика сериализуется в JSON и отправляется в ответе с HTTP-статусом 200 (OK).
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий метрику в формате JSON в теле запроса
func (ms *MetricsServer) UpdateMetric(res http.ResponseWriter, req *http.Request) {
	ms.syncMutex.Lock()
	defer ms.syncMutex.Unlock()

	ctx := context.TODO()

	res.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	mTemp := &service.Metrics{}
	var bufJSON bytes.Buffer

	_, err := bufJSON.ReadFrom(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	if err := json.Unmarshal(bufJSON.Bytes(), mTemp); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := ms.MetricStorage.Save(ctx, *mTemp); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if mTemp, err = ms.MetricStorage.Get(ctx, mTemp.ID); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	bufResp, err := json.Marshal(mTemp)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)

	res.Write(bufResp)

}

// ExtractMetric обрабатывает HTTP-запросы на получение метрик через JSON.
//
// Метод принимает метрику в формате JSON, проверяет её корректность,
// извлекает метрику из хранилища и возвращает её в ответе.
//
// Логика работы:
//  1. Проверяется, что метод запроса является POST. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  2. Тело запроса считывается и десериализуется в объект Metrics.
//     Если данные некорректны, возвращается ошибка 404 (Not Found).
//  3. Проверяется соответствие типа метрики (MType) между запрошенной и полученной из хранилища.
//     Если типы не совпадают, возвращается ошибка 404.
//  4. Метрика извлекается из хранилища по её ID. Если метрика не найдена, возвращается ошибка 404.
//  5. Обновленная метрика сериализуется в JSON и отправляется в ответе с HTTP-статусом 200 (OK).
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий метрику в формате JSON в теле запроса.
func (ms *MetricsServer) ExtractMetric(res http.ResponseWriter, req *http.Request) {
	ctx := context.TODO()
	res.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	mTemp := &service.Metrics{}
	var bufJSON bytes.Buffer

	_, err := bufJSON.ReadFrom(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	defer req.Body.Close()
	if err := json.Unmarshal(bufJSON.Bytes(), mTemp); err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	mType := mTemp.MType

	if mTemp, err = ms.MetricStorage.Get(ctx, mTemp.ID); err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if mTemp.MType != mType {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	bufResp, err := json.Marshal(mTemp)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.WriteHeader(http.StatusOK)

	res.Write(bufResp)
}

// HandleGetMetric обрабатывает HTTP-запросы на получение значения метрики.
//
// Метод извлекает тип и имя метрики из параметров запроса, проверяет их корректность,
// и возвращает значение метрики в ответе в текстовом формате (text/html).
//
// Логика работы:
//  1. Проверяется, что метод запроса является GET. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  2. Извлекается тип метрики ("type") из URL-параметров. Если тип пустой, возвращается ошибка 404 (Not Found).
//  3. Извлекается имя метрики ("name") из URL-параметров и запрашивается метрика из хранилища.
//     Если метрика не найдена, возвращается ошибка 404.
//  4. В зависимости от типа метрики (gauge или counter) возвращается её значение:
//     - Для "gauge" возвращается значение поля Value.
//     - Для "counter" возвращается значение поля Delta.
//  5. Значение метрики записывается в тело HTTP-ответа в текстовом формате.
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий параметры пути "type" и "name".
func (ms *MetricsServer) HandleGetMetric(res http.ResponseWriter, req *http.Request) {
	ctx := context.TODO()
	res.Header().Set("Content-Type", "text/html")
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	t := chi.URLParam(req, "type")
	if len(t) == 0 {
		http.Error(res, "Incorrect type!", http.StatusNotFound)
		return
	}
	n := chi.URLParam(req, "name")
	mTemp, err := ms.MetricStorage.Get(ctx, n)
	if err != nil {
		http.Error(res, "Metric not found!", http.StatusNotFound)
		return
	}
	switch mTemp.MType {
	case service.GaugeMetric:
		value := mTemp.Value
		logging.Logg.Info("res", "%v", *value)
		fmt.Fprintf(res, "%v", *value)

	case service.CounterMetric:
		value := mTemp.Delta
		logging.Logg.Info("res", "%v", *value)
		fmt.Fprintf(res, "%v", *value)
	}
}

// HandleGetAllMetrics обрабатывает HTTP-запросы на получение всех метрик в виде HTML-страницы.
//
// Метод извлекает все метрики из хранилища, записывает их в HTML-шаблон и возвращает результат клиенту.
//
// Логика работы:
//  1. Проверяется, что метод запроса является GET. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  2. Загружается HTML-шаблон из файла "cmd/server/static/index.html.tmpl".
//     Если шаблон не может быть загружен, возвращается ошибка 500 (Internal Server Error).
//  3. Все метрики извлекаются из хранилища с помощью метода List.
//     Если возникает ошибка при получении метрик, возвращается ошибка 500.
//  4. Метрики передаются в HTML-шаблон, который записывается в тело HTTP-ответа.
//     Если при записи возникает ошибка, возвращается ошибка 500.
//  5. В случае успеха возвращается HTTP-статус 200 (OK) с HTML-страницей.
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий запрос на получение всех метрик.
func (ms *MetricsServer) HandleGetAllMetrics(res http.ResponseWriter, req *http.Request) {
	ctx := context.TODO()
	res.Header().Set("Content-Type", "text/html")

	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("cmd/server/static/index.html.tmpl")
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	mtrx, err := ms.MetricStorage.List(ctx)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(res, *mtrx)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// CheckDBConnect обрабатывает HTTP-запросы на проверку подключения к базе данных.
//
// Метод проверяет состояние подключения к хранилищу метрик и возвращает результат проверки.
//
// Логика работы:
//  1. Проверяется, что метод запроса является GET. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  2. Вызывается метод CheckStorage хранилища метрик для проверки состояния подключения.
//     Если возникает ошибка, возвращается ошибка 500 (Internal Server Error) с сообщением "database connection failed".
//  3. Если проверка прошла успешно, возвращается HTTP-статус 200 (OK) и текст "Status OK".
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий запрос на проверку подключения к базе данных.
func (ms *MetricsServer) CheckDBConnect(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	if err := ms.MetricStorage.CheckStorage(); err != nil {
		http.Error(res, "database connection failed", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Status OK"))
}

// UpdateBatch обрабатывает HTTP-запросы на пакетное обновление метрик.
//
// Метод принимает массив метрик в формате JSON, проверяет их корректность,
// сохраняет в хранилище и возвращает обновленный список всех метрик в ответе.
//
// Логика работы:
//  1. Используется мьютекс для обеспечения потокобезопасности при работе с хранилищем.
//  2. Проверяется, что метод запроса является POST. Если нет, возвращается ошибка 405 (Method Not Allowed).
//  3. Тело запроса считывается и десериализуется в массив метрик.
//     Если данные некорректны или превышают допустимый размер, возвращаются соответствующие ошибки:
//     - 413 Request Entity Too Large — если размер данных превышает лимит.
//     - 400 Bad Request — если данные некорректны.
//  4. Массив метрик сохраняется в хранилище. Если сохранение завершается ошибкой, возвращается ошибка 400.
//  5. После сохранения извлекается полный список метрик из хранилища.
//  6. Список метрик сериализуется в JSON и отправляется в ответе с HTTP-статусом 200 (OK).
//  7. Если в конфигурации сервера задан ключ (ms.Config.Key), вычисляется хеш SHA-256 для ответа
//     и добавляется в заголовок "HashSHA256".
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий массив метрик в формате JSON в теле запроса.
func (ms *MetricsServer) UpdateBatch(res http.ResponseWriter, req *http.Request) {
	ms.syncMutex.Lock()
	defer ms.syncMutex.Unlock()

	ctx := context.TODO()

	res.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	var allMtrx *map[string]service.Metrics
	var mTemp []service.Metrics

	if err := ReadAndUnmarshal(req, &mTemp); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			res.WriteHeader(http.StatusRequestEntityTooLarge)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	err := ms.MetricStorage.SaveAll(ctx, &mTemp)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	allMtrx, err = ms.MetricStorage.List(ctx)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	bufResp, err := json.Marshal(allMtrx)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(ms.Config.Key) > 0 {
		signBuf := bufResp
		signBuf = append(signBuf, ',')
		signBuf = append(signBuf, ms.Config.Key...)

		sign := sha256.Sum256(signBuf)
		res.Header().Set("HashSHA256", hex.EncodeToString(sign[:]))
	}
	res.WriteHeader(http.StatusOK)

	res.Write(bufResp)

}

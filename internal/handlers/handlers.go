// Package handlers предоставляет HTTP-обработчики для работы с метриками.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// HandlePutGaugeMetric обрабатывает HTTP-запросы на сохранение метрики типа "gauge".
//
// Метод извлекает имя метрики и её значение из параметров запроса, проверяет
// их корректность, и сохраняет метрику в хранилище.
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
// Метод извлекает имя метрики и её значение из параметров запроса, проверяет
// их корректность, и сохраняет метрику в хранилище.
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
// Метод принимает метрику в формате JSON, проверяет её корректность, сохраняет
// в хранилище и возвращает обновленную метрику в ответе.
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
// Метод принимает метрику в формате JSON, проверяет её корректность, извлекает
// метрику из хранилища и возвращает её в ответе.
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
// Метод извлекает тип и имя метрики из параметров запроса, проверяет их
// корректность, и возвращает значение метрики в ответе в текстовом формате
// (text/html).
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
// Метод извлекает все метрики из хранилища, записывает их в HTML-шаблон и
// возвращает результат клиенту.
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
// Метод проверяет состояние подключения к хранилищу метрик и возвращает результат проверки.
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
// Метод принимает зашифрованный массив метрик в формате JSON, проверяет их корректность,
// сохраняет в хранилище и возвращает обновленный список всех метрик в ответе.
//
// Параметры:
// - res: HTTP-ответ, который будет отправлен клиенту.
// - req: HTTP-запрос, содержащий массив метрик в формате JSON в теле запроса.
func (ms *MetricsServer) UpdateBatch(res http.ResponseWriter, req *http.Request) {
	ms.syncMutex.Lock()
	defer ms.syncMutex.Unlock()

	ctx := context.TODO()

	res.Header().Set("Content-Type", "application/json")

	if err := ms.checkPostMethod(req); err != nil {
		http.Error(res, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	body, err := ms.readRequestBody(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	privateKey, err := ms.loadPrivateKey()
	if err != nil {
		logging.Logg.Error("Failed to load private key: %v", err)
		http.Error(res, "Failed to load private key", http.StatusInternalServerError)
		return
	}
	decryptedData, err := ms.decryptData(body, privateKey)
	if err != nil {
		logging.Logg.Error("Failed to decrypt data: %v", err)
		http.Error(res, "Failed to decrypt data", http.StatusInternalServerError)
		return
	}

	metrics, err := ms.parseMetrics(decryptedData)
	if err != nil {
		logging.Logg.Error("Failed to parse metrics: %v", err)
		http.Error(res, "Failed to parse metrics", http.StatusBadRequest)
		return
	}

	if err := ms.saveMetrics(ctx, metrics); err != nil {
		http.Error(res, "Failed to save metrics", http.StatusBadRequest)
		return
	}

	allMetrics, err := ms.getAllMetrics(ctx)
	if err != nil {
		http.Error(res, "Failed to retrieve metrics", http.StatusBadRequest)
		return
	}

	response, hash, err := ms.prepareResponse(allMetrics, ms.Config.Key)
	if err != nil {
		http.Error(res, "Failed to prepare response", http.StatusBadRequest)
		return
	}

	if hash != "" {
		res.Header().Set("HashSHA256", hash)
	}

	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

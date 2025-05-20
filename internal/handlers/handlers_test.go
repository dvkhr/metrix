package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/mocks"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlePutGaugeMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockMetricStorage(ctrl)

	server := &MetricsServer{
		MetricStorage: mockStorage,
	}

	t.Run("Successful POST Request", func(t *testing.T) {

		metricName := "test_metric"
		metricValue := 42.0
		gaugeValue := service.GaugeMetricValue(metricValue)
		mockStorage.EXPECT().
			Save(gomock.Any(), service.Metrics{
				ID:    metricName,
				MType: service.GaugeMetric,
				Value: &gaugeValue,
			}).
			Return(nil)

		mockStorage.EXPECT().
			Get(gomock.Any(), metricName).
			Return(&service.Metrics{
				ID:    metricName,
				MType: service.GaugeMetric,
				Value: &gaugeValue,
			}, nil)

		router := chi.NewRouter()
		router.Post("/update/gauge/{name}/{value}", server.HandlePutGaugeMetric)

		reqPath := fmt.Sprintf("/update/gauge/%s/%f", metricName, metricValue)
		req := httptest.NewRequest(http.MethodPost, reqPath, nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("Unsupported HTTP Method", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/gauge/{name}/{value}", server.HandlePutGaugeMetric)

		req := httptest.NewRequest(http.MethodGet, "/update/gauge/test_metric/42", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	})

	t.Run("Invalid Metric Name", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/gauge/{name}/{value}", server.HandlePutGaugeMetric)

		req := httptest.NewRequest(http.MethodPost, "/update/gauge//42", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
		assert.Contains(t, res.Body.String(), "Incorrect name!")
	})

	t.Run("Invalid Metric Value", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/gauge/{name}/{value}", server.HandlePutGaugeMetric)

		req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/invalid_value", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "Incorrect value!")
	})
}

func TestHandlePutCounterMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockMetricStorage(ctrl)

	server := &MetricsServer{
		MetricStorage: mockStorage,
	}

	t.Run("Successful POST Request", func(t *testing.T) {

		metricName := "test_metric"
		metricValue := int64(42)
		counterValue := service.CounterMetricValue(metricValue)
		mockStorage.EXPECT().
			Save(gomock.Any(), service.Metrics{
				ID:    metricName,
				MType: service.CounterMetric,
				Delta: &counterValue,
			}).
			Return(nil)

		mockStorage.EXPECT().
			Get(gomock.Any(), metricName).
			Return(&service.Metrics{
				ID:    metricName,
				MType: service.CounterMetric,
				Delta: &counterValue,
			}, nil)

		router := chi.NewRouter()
		router.Post("/update/counter/{name}/{value}", server.HandlePutCounterMetric)

		reqPath := fmt.Sprintf("/update/counter/%s/%d", metricName, metricValue)
		req := httptest.NewRequest(http.MethodPost, reqPath, nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("Unsupported HTTP Method", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/counter/{name}/{value}", server.HandlePutCounterMetric)

		req := httptest.NewRequest(http.MethodGet, "/update/counter/test_metric/42", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	})

	t.Run("Invalid Metric Name", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/counter/{name}/{value}", server.HandlePutCounterMetric)

		req := httptest.NewRequest(http.MethodPost, "/update/counter//42", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		body := res.Body.String()
		assert.Contains(t, body, "Incorrect name!")
	})

	t.Run("Invalid Metric Value", func(t *testing.T) {

		router := chi.NewRouter()
		router.Post("/update/counter/{name}/{value}", server.HandlePutCounterMetric)

		req := httptest.NewRequest(http.MethodPost, "/update/counter/test_metric/invalid_value", nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		body := res.Body.String()
		assert.Contains(t, body, "Incorrect value!")
	})
}
func TestUpdateBatch(t *testing.T) {
	t.Run("Successful POST Request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := mocks.NewMockMetricStorage(ctrl)

		server := &MetricsServer{
			MetricStorage: mockStorage,
			Config:        config.ConfigServ{Key: "test_key"},
		}

		counterValue := service.CounterMetricValue(100)
		gaugeValue := service.GaugeMetricValue(42)

		metrics := []service.Metrics{
			{ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
			{ID: "metric_2", MType: service.CounterMetric, Delta: &counterValue},
		}

		mockStorage.EXPECT().
			SaveAll(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		metricsMap := map[string]service.Metrics{
			"metric_1": {ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
			"metric_2": {ID: "metric_2", MType: service.CounterMetric, Delta: &counterValue},
		}

		mockStorage.EXPECT().
			List(gomock.Any()).
			Return(&metricsMap, nil).
			Times(1)

		jsonData, _ := json.Marshal(metrics)
		req := httptest.NewRequest(http.MethodPost, "/update/batch", bytes.NewReader(jsonData))
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		var response map[string]service.Metrics
		err := json.Unmarshal(res.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "metric_1")
		assert.Contains(t, response, "metric_2")
	})

	t.Run("Unsupported HTTP Method", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := mocks.NewMockMetricStorage(ctrl)

		server := &MetricsServer{
			MetricStorage: mockStorage,
			Config:        config.ConfigServ{Key: "test_key"},
		}

		req := httptest.NewRequest(http.MethodGet, "/update/batch", nil)
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		assert.Equal(t, http.StatusMethodNotAllowed, res.Code)

		body := res.Body.String()
		assert.Contains(t, body, "only POST requests are allowed")
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		if err := logging.InitTestLogger(); err != nil {
			fmt.Printf("Failed to initialize test logger: %v\n", err)
			return
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := mocks.NewMockMetricStorage(ctrl)

		server := &MetricsServer{
			MetricStorage: mockStorage,
			Config: config.ConfigServ{
				Key: "test_key", // Подпись включена
			},
		}

		// Отправляем некорректный JSON
		req := httptest.NewRequest(http.MethodPost, "/update/batch", bytes.NewReader([]byte("invalid_json")))
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		// Проверяем, что статус равен 400
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("SaveAll Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := mocks.NewMockMetricStorage(ctrl)

		server := &MetricsServer{
			MetricStorage: mockStorage,
			Config:        config.ConfigServ{Key: "test_key"},
		}

		gaugeValue := service.GaugeMetricValue(42)

		metrics := []service.Metrics{
			{ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
		}

		mockStorage.EXPECT().
			SaveAll(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("storage error")).
			Times(1)

		jsonData, _ := json.Marshal(metrics)
		req := httptest.NewRequest(http.MethodPost, "/update/batch", bytes.NewReader(jsonData))
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("List Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := mocks.NewMockMetricStorage(ctrl)

		server := &MetricsServer{
			MetricStorage: mockStorage,
			Config:        config.ConfigServ{Key: "test_key"},
		}

		gaugeValue := service.GaugeMetricValue(42)

		metrics := []service.Metrics{
			{ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
		}

		mockStorage.EXPECT().
			SaveAll(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		mockStorage.EXPECT().
			List(gomock.Any()).
			Return(nil, fmt.Errorf("list error")).
			Times(1)

		jsonData, _ := json.Marshal(metrics)
		req := httptest.NewRequest(http.MethodPost, "/update/batch", bytes.NewReader(jsonData))
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
}

// go test -bench=. -memprofile=mem.pprof
// go tool pprof -http=":9090" handlers.test mem.pprof

func BenchmarkUpdateBatch(b *testing.B) {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockMetricStorage(ctrl)

	server := &MetricsServer{
		MetricStorage: mockStorage,
		Config:        config.ConfigServ{Key: "test_key"}, // Ключ для подписи
	}

	counterValue := service.CounterMetricValue(100)
	gaugeValue := service.GaugeMetricValue(42.0)
	metrics := []service.Metrics{
		{ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
		{ID: "metric_2", MType: service.CounterMetric, Delta: &counterValue},
	}

	mockStorage.EXPECT().
		SaveAll(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	mockStorage.EXPECT().
		List(gomock.Any()).
		Return(&map[string]service.Metrics{
			"metric_1": {ID: "metric_1", MType: service.GaugeMetric, Value: &gaugeValue},
			"metric_2": {ID: "metric_2", MType: service.CounterMetric, Delta: &counterValue},
		}, nil).
		AnyTimes()

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		b.Fatalf("Failed to marshal metrics: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/batch", bytes.NewReader(jsonData))
		res := httptest.NewRecorder()

		server.UpdateBatch(res, req)

		if res.Code != http.StatusOK {
			b.Fatalf("Unexpected status code: %d", res.Code)
		}
	}
}

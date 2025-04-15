package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

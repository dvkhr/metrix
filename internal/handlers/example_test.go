package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/handlers"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/routes"
	"github.com/dvkhr/metrix.git/internal/service"
)

func init() {
	// Инициализация логгера

	logging.Logg = logging.NewLogger("none", "text", "json", "concole", "logs/2006-01-02.log")
	if logging.Logg == nil {
		fmt.Println("Failed to initialize logger")
		//os.Exit(1)
	}
}

// ExampleHandlePutGaugeMetric демонстрирует работу с эндпоинтом /update/gauge/{name}/{value}.
func ExampleHandlePutGaugeMetric() {

	cfg := config.ConfigServ{
		Address: ":8080",
	}

	metricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		panic("Failed to initialize MetricsServer: " + err.Error())
	}

	r := routes.SetupRoutes(cfg, metricServer)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	req, err := http.NewRequest("POST", testServer.URL+"/update/gauge/test_metric/42.0", nil)
	if err != nil {
		panic("Failed to create request: " + err.Error())
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Metric updated successfully")
	} else {
		fmt.Println("Failed to update metric")
	}

	// Output:
	// Metric updated successfully
}

// ExampleHandlePutCounterMetric демонстрирует работу с эндпоинтом /update/counter/{name}/{value}.
func ExampleHandlePutCounterMetric() {
	cfg := config.ConfigServ{
		Address: ":8080",
	}

	metricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		panic("Failed to initialize MetricsServer: " + err.Error())
	}
	r := routes.SetupRoutes(cfg, metricServer)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	req, err := http.NewRequest("POST", testServer.URL+"/update/counter/test_counter/42", nil)
	if err != nil {
		panic("Failed to create request: " + err.Error())
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Counter updated successfully")
	} else {
		fmt.Println("Failed to update counter")
	}

	// Output:
	// Counter updated successfully
}

// ExampleUpdateMetric демонстрирует работу с эндпоинтом /update/.
func ExampleUpdateMetric() {
	cfg := config.ConfigServ{
		Address: ":8080",
	}

	metricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		panic("Failed to initialize MetricsServer: " + err.Error())
	}

	r := routes.SetupRoutes(cfg, metricServer)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	gaugeValue := service.GaugeMetricValue(42)
	metric := service.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: &gaugeValue,
	}
	jsonBody, err := json.Marshal(metric)
	if err != nil {
		panic("Failed to marshal JSON body: " + err.Error())
	}

	req, err := http.NewRequest("POST", testServer.URL+"/update/", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic("Failed to create request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Metric updated successfully")
	} else {
		fmt.Printf("Failed to update metric. Status code: %d\n", resp.StatusCode)
	}

	// Output:
	// Metric updated successfully
}

// ExampleExtractMetric демонстрирует работу с эндпоинтом /value/.
func ExampleExtractMetric() {
	cfg := config.ConfigServ{
		Address: ":8080",
	}

	metricServer, err := handlers.NewMetricsServer(cfg)
	if err != nil {
		panic("Failed to initialize MetricsServer: " + err.Error())
	}

	r := routes.SetupRoutes(cfg, metricServer)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	gaugeValue := service.GaugeMetricValue(42)
	metricToSave := service.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	ctx := context.TODO()
	if err := metricServer.MetricStorage.Save(ctx, metricToSave); err != nil {
		panic("Failed to save metric: " + err.Error())
	}

	metricToExtract := service.Metrics{
		ID:    "test_metric",
		MType: "gauge",
	}
	jsonBody, err := json.Marshal(metricToExtract)
	if err != nil {
		panic("Failed to marshal JSON body: " + err.Error())
	}

	req, err := http.NewRequest("POST", testServer.URL+"/value/", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic("Failed to create request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Metric extracted successfully")
	} else {
		fmt.Printf("Failed to extract metric. Status code: %d\n", resp.StatusCode)
	}

	// Output:
	// Metric extracted successfully
}

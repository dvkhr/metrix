package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/go-chi/chi/v5"
)

type MetricStorage interface {
	PutGaugeMetric(metricName string, metricValue metric.GaugeMetricValue) error
	PutCounterMetric(metricName string, metricValue metric.CounterMetricValue) error
	GetGaugeMetric(metricName string) (metric.GaugeMetricValue, error)
	GetCounterMetric(metricName string) (metric.CounterMetricValue, error)
	AllMetrics() (*map[string]interface{}, error)

	NewMemStorage()
}

type MetricsServer struct {
	MetricStorage MetricStorage
}

func NewMetricsServer(MetricStorage MetricStorage) *MetricsServer {
	MetricStorage.NewMemStorage()
	return &MetricsServer{MetricStorage: MetricStorage}
}

func (ms *MetricsServer) HandlePutGaugeMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	n := req.PathValue("name")
	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}

	/*Debug output*/
	sv, _ := ms.MetricStorage.GetGaugeMetric(n)
	fmt.Printf("BEFORE: Metric %s has value %g\n", n, sv)

	v, err := strconv.ParseFloat(req.PathValue("value"), 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	ms.MetricStorage.PutGaugeMetric(req.PathValue("name"), metric.GaugeMetricValue(v))

	/*Debug output*/
	sv, _ = ms.MetricStorage.GetGaugeMetric(n)
	fmt.Printf("AFTER: Metric %s has value %v\n", n, sv)
	res.Write([]byte("Done!"))
}

func (ms *MetricsServer) HandlePutCounterMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	n := req.PathValue("name")

	/*Debug output*/
	sv, _ := ms.MetricStorage.GetCounterMetric(n)
	fmt.Printf("BEFORE: Metric %s has value %v\n", n, sv)

	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}
	v, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	ms.MetricStorage.PutCounterMetric(req.PathValue("name"), metric.CounterMetricValue(v))

	/*Debug output*/
	sv, _ = ms.MetricStorage.GetCounterMetric(n)
	fmt.Printf("AFTER: Metric %s has value %v\n", n, sv)
	res.Write([]byte("Done!"))
}

func (ms *MetricsServer) IncorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

func (ms *MetricsServer) NotfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
}

func (ms *MetricsServer) HandleGetMetric(res http.ResponseWriter, req *http.Request) {
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
	switch metric.Metric(t) {
	case metric.GaugeMetric:
		value, err := ms.MetricStorage.GetGaugeMetric(n)
		if err != nil {
			http.Error(res, "Metric not found!", http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%v", value)
	case metric.CounterMetric:
		value, err := ms.MetricStorage.GetCounterMetric(n)
		if err != nil {
			http.Error(res, "Metric not found!", http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%v", value)
	default:
		http.Error(res, "Metric not found!", http.StatusNotFound)
	}
}
func (ms *MetricsServer) HandleGetAllMetrics(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("index.html.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	mtrx, err := ms.MetricStorage.AllMetrics()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(res, *mtrx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

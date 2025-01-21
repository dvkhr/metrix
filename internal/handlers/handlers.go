package handlers

import (
	"fmt"
	"net/http"
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
	MetricStrings() ([]string, error)
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
	v, err := strconv.ParseFloat(req.PathValue("value"), 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	ms.MetricStorage.PutGaugeMetric(req.PathValue("name"), metric.GaugeMetricValue(v))
	res.WriteHeader(http.StatusOK)

}

func (ms *MetricsServer) HandlePutCounterMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
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
	ms.MetricStorage.PutCounterMetric(req.PathValue("name"), metric.CounterMetricValue(v))
	res.WriteHeader(http.StatusOK)
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
	const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Metrics</title>
</head>
<body>
	<table border="1">
		<tr>
			<th>Metric</th>
		</tr>
		{{range .}}
			<tr>
				<td>{{.}}</td>
			</tr>
		{{end}}
	</table>
</body>
</html>
	`

	metricStrings, err := ms.MetricStorage.MetricStrings()
	if err != nil {
		http.Error(res, "Metrics not found!", http.StatusNotFound)
	}

	tmpl, err := template.New("allMetrics").Parse(htmlTemplate)
	if err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(res, "allMetrics", metricStrings)
	res.WriteHeader(http.StatusOK)
}

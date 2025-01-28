package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/go-chi/chi/v5"
)

type MetricStorage interface {
	PutMetric(mt metric.Metrics) error
	GetMetric(metricName string) (*metric.Metrics, error)
	AllMetrics() (*map[string]metric.Metrics, error)
	NewMemStorage()
}

type MetricsServer struct {
	MetricStorage MetricStorage
}

func NewMetricsServer(MetricStorage MetricStorage) *MetricsServer {
	MetricStorage.NewMemStorage()
	return &MetricsServer{MetricStorage: MetricStorage}
}

func (ms *MetricsServer) IncorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

func (ms *MetricsServer) NotfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
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
	mTemp := &metric.Metrics{}
	mTemp.ID = n

	vtemp := metric.GaugeMetricValue(v)
	mTemp.Value = &vtemp
	mTemp.MType = metric.GaugeMetric

	ms.MetricStorage.PutMetric(*mTemp)
	res.WriteHeader(http.StatusOK)
	ms.MetricStorage.GetMetric(req.PathValue("name"))
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
	mTemp := &metric.Metrics{}
	mTemp.ID = n

	vtemp := metric.CounterMetricValue(v)
	mTemp.Delta = &vtemp
	mTemp.MType = metric.CounterMetric

	ms.MetricStorage.PutMetric(*mTemp)
	res.WriteHeader(http.StatusOK)
}

func (ms *MetricsServer) HandlePutMetricJSON(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	mTemp := &metric.Metrics{}
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
	if err := ms.MetricStorage.PutMetric(*mTemp); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if mTemp, err = ms.MetricStorage.GetMetric(mTemp.ID); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	bufResp, err := json.Marshal(mTemp)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Write(bufResp)
	res.WriteHeader(http.StatusOK)
}

func (ms *MetricsServer) HandleGetMetricJSON(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	mTemp := &metric.Metrics{}
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

	if mTemp, err = ms.MetricStorage.GetMetric(mTemp.ID); err != nil {
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

	res.Write(bufResp)
	res.WriteHeader(http.StatusOK)
}

func (ms *MetricsServer) HandleGetMetric(res http.ResponseWriter, req *http.Request) {
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
	//mTemp := &metric.Metrics{}
	mTemp, err := ms.MetricStorage.GetMetric(n)
	if err != nil {
		http.Error(res, "Metric not found!", http.StatusNotFound)
		return
	}
	switch mTemp.MType {
	case metric.GaugeMetric:
		value := mTemp.Value
		fmt.Fprintf(res, "%v", *value)
	case metric.CounterMetric:
		value := mTemp.Delta
		fmt.Fprintf(res, "%v", *value)
	}
}

func (ms *MetricsServer) HandleGetAllMetrics(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	tmpl, err := template.ParseFiles("cmd/server/static/index.html.tmpl")
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
	res.WriteHeader(http.StatusOK)
}

func DumpMetrics(ms *MetricsServer, wr io.Writer) error {

	mtrx, err := ms.MetricStorage.AllMetrics()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(mtrx, "", "  ")
	if err != nil {
		return err
	}
	_, err = wr.Write(data)
	return err
}
func RestoreMetrics(ms *MetricsServer, rd io.Reader) error {
	var data []byte

	data, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	stor, err := ms.MetricStorage.AllMetrics()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, stor)

	return err
}

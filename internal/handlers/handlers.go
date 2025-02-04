package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/logger"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type MetricStorage interface {
	Save(mt service.Metrics) error
	Get(metricName string) (*service.Metrics, error)
	List() (*map[string]service.Metrics, error)
	NewMemStorage()
}

type MetricsServer struct {
	MetricStorage MetricStorage
	Config        config.ConfigServ
	Sync          bool
	mutex         sync.Mutex
	DB            *sql.DB
}

func NewMetricsServer(MetricStorage MetricStorage, Config config.ConfigServ) *MetricsServer {
	MetricStorage.NewMemStorage()
	var s bool
	if Config.StoreInterval == 0*time.Second {
		s = true
	} else {
		s = false
	}
	return &MetricsServer{MetricStorage: MetricStorage, Config: Config, Sync: s}
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
	mTemp := &service.Metrics{}
	mTemp.ID = n

	vtemp := service.GaugeMetricValue(v)
	mTemp.Value = &vtemp
	mTemp.MType = service.GaugeMetric

	ms.MetricStorage.Save(*mTemp)
	res.WriteHeader(http.StatusOK)
	ms.MetricStorage.Get(req.PathValue("name"))
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
	mTemp := &service.Metrics{}
	mTemp.ID = n

	vtemp := service.CounterMetricValue(v)
	mTemp.Delta = &vtemp
	mTemp.MType = service.CounterMetric

	ms.MetricStorage.Save(*mTemp)
	res.WriteHeader(http.StatusOK)
}

func (ms *MetricsServer) UpdateMetric(res http.ResponseWriter, req *http.Request) {
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
	if err := ms.MetricStorage.Save(*mTemp); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if mTemp, err = ms.MetricStorage.Get(mTemp.ID); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	bufResp, err := json.Marshal(mTemp)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if ms.Sync {
		ms.DumpMetrics()
	}

	res.Write(bufResp)
	res.WriteHeader(http.StatusOK)

}

func (ms *MetricsServer) ExtractMetric(res http.ResponseWriter, req *http.Request) {
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

	if mTemp, err = ms.MetricStorage.Get(mTemp.ID); err != nil {
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
	mTemp, err := ms.MetricStorage.Get(n)
	if err != nil {
		http.Error(res, "Metric not found!", http.StatusNotFound)
		return
	}
	switch mTemp.MType {
	case service.GaugeMetric:
		value := mTemp.Value
		fmt.Fprintf(res, "%v", *value)
	case service.CounterMetric:
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
	mtrx, err := ms.MetricStorage.List()
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
func (ms *MetricsServer) DumpMetrics() {
	file, err := os.OpenFile(ms.Config.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Sugar.Errorln("unable open file", "file", ms.Config.FileStoragePath, "error", err)
	}
	ms.mutex.Lock()
	err = service.DumpMetrics(ms.MetricStorage, file)
	if err != nil {
		logger.Sugar.Errorln("unable dump metrics", "error", err)
	}
	err = file.Sync()
	if err != nil {
		logger.Sugar.Errorln("unable sync file", "error", err)
	}
	ms.mutex.Unlock()
	err = file.Close()
	if err != nil {
		logger.Sugar.Errorln("unable close file", "error", err)
	}

	logger.Sugar.Infoln("metrics dumped")
}
func (ms *MetricsServer) LoadMetrics() {
	file, err := os.OpenFile(ms.Config.FileStoragePath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Sugar.Errorln("unable open file", "file", ms.Config.FileStoragePath, "error", err)
	} else {
		err = service.RestoreMetrics(ms.MetricStorage, file)
		if err != nil {
			logger.Sugar.Errorln("unable to restore metrics", "error", err)
		} else {
			logger.Sugar.Infoln("metrics restored")
		}
		file.Close()
		if err != nil {
			logger.Sugar.Errorln("unable close file", "error", err)
		}
	}
}
func (ms *MetricsServer) CheckDBConnect(res http.ResponseWriter, req *http.Request) {
	err := ms.DB.Ping()
	if err != nil {
		logger.Sugar.Errorln("database connection failed", "error", err)
		http.Error(res, "database connection failed", http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infoln("connection to the database has been successfully")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Status OK"))
}

package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

type MetricsServer struct {
	MetricStorage MetricStorage
	Config        config.ConfigServ
	syncMutex     sync.Mutex
}

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

func (ms *MetricsServer) IncorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

func (ms *MetricsServer) NotfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
}

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

	var bufJSON bytes.Buffer
	_, err := bufJSON.ReadFrom(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	if bufJSON.Len() > 0 {
		if err := json.Unmarshal(bufJSON.Bytes(), &mTemp); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := ms.MetricStorage.SaveAll(ctx, &mTemp); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if allMtrx, err = ms.MetricStorage.List(ctx); err != nil {
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
		req.Header.Set("HashSHA256", hex.EncodeToString(sign[:]))
	}
	res.WriteHeader(http.StatusOK)

	res.Write(bufResp)

}

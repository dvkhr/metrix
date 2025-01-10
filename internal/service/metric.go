package metric

import (
	"errors"
	"math/rand"
	"runtime"

	"github.com/dvkhr/metrix.git/internal/storage"
)

type Metric string

const (
	GaugeMetric   Metric = "gauge"
	CounterMetric Metric = "counter"
)

var ErrUninitializedStorage = errors.New("storage is not initialized")
var ErrInvalidMetricName = errors.New("invalid metric name")
var ErrUnkonownMetric = errors.New("unknown metric")

type GaugeMetricValue float64
type CounterMetricValue int64

type MetricStorage interface {
	PutGaugeMetric(metricName string, metricValue storage.GaugeMetricValue) error
	PutCounterMetric(metricName string, metricValue storage.CounterMetricValue) error
}

func CollectMetrics(ms MetricStorage) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	ms.PutGaugeMetric("Alloc", storage.GaugeMetricValue(rtm.Alloc))
	ms.PutGaugeMetric("BuckHashSys", storage.GaugeMetricValue(rtm.BuckHashSys))
	ms.PutGaugeMetric("Frees", storage.GaugeMetricValue(rtm.Frees))
	ms.PutGaugeMetric("GCCPUFraction", storage.GaugeMetricValue(rtm.GCCPUFraction))
	ms.PutGaugeMetric("GCSys", storage.GaugeMetricValue(rtm.GCSys))
	ms.PutGaugeMetric("HeapAlloc", storage.GaugeMetricValue(rtm.HeapAlloc))
	ms.PutGaugeMetric("HeapIdle", storage.GaugeMetricValue(rtm.HeapIdle))
	ms.PutGaugeMetric("HeapInuse", storage.GaugeMetricValue(rtm.HeapInuse))
	ms.PutGaugeMetric("HeapObjects", storage.GaugeMetricValue(rtm.HeapObjects))
	ms.PutGaugeMetric("HeapReleased", storage.GaugeMetricValue(rtm.HeapReleased))
	ms.PutGaugeMetric("HeapSys", storage.GaugeMetricValue(rtm.HeapSys))
	ms.PutGaugeMetric("LastGC", storage.GaugeMetricValue(rtm.LastGC))
	ms.PutGaugeMetric("Lookups", storage.GaugeMetricValue(rtm.Lookups))
	ms.PutGaugeMetric("MCacheInuse", storage.GaugeMetricValue(rtm.MCacheInuse))
	ms.PutGaugeMetric("MCacheSys", storage.GaugeMetricValue(rtm.MCacheSys))
	ms.PutGaugeMetric("MSpanInuse", storage.GaugeMetricValue(rtm.MSpanInuse))
	ms.PutGaugeMetric("MSpanSys", storage.GaugeMetricValue(rtm.MSpanSys))
	ms.PutGaugeMetric("Mallocs", storage.GaugeMetricValue(rtm.Mallocs))
	ms.PutGaugeMetric("NextGC", storage.GaugeMetricValue(rtm.NextGC))
	ms.PutGaugeMetric("NumForcedGC", storage.GaugeMetricValue(rtm.NumForcedGC))
	ms.PutGaugeMetric("NumGC", storage.GaugeMetricValue(rtm.NumGC))
	ms.PutGaugeMetric("OtherSys", storage.GaugeMetricValue(rtm.OtherSys))
	ms.PutGaugeMetric("PauseTotalNs", storage.GaugeMetricValue(rtm.PauseTotalNs))
	ms.PutGaugeMetric("StackInuse", storage.GaugeMetricValue(rtm.StackInuse))
	ms.PutGaugeMetric("StackSys", storage.GaugeMetricValue(rtm.StackSys))
	ms.PutGaugeMetric("Sys", storage.GaugeMetricValue(rtm.Sys))
	ms.PutGaugeMetric("TotalAlloc", storage.GaugeMetricValue(rtm.TotalAlloc))
	ms.PutGaugeMetric("RandomValue", storage.GaugeMetricValue(rand.Float64()))
	ms.PutCounterMetric("PollCount", storage.CounterMetricValue(rand.Float64()))
}

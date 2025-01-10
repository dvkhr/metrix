package metric

import (
	"errors"
	"math/rand"
	"runtime"
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
	PutGaugeMetric(metricName string, metricValue GaugeMetricValue) error
	PutCounterMetric(metricName string, metricValue CounterMetricValue) error
}

func CollectMetrics(ms MetricStorage) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	ms.PutGaugeMetric("Alloc", GaugeMetricValue(rtm.Alloc))
	ms.PutGaugeMetric("BuckHashSys", GaugeMetricValue(rtm.BuckHashSys))
	ms.PutGaugeMetric("Frees", GaugeMetricValue(rtm.Frees))
	ms.PutGaugeMetric("GCCPUFraction", GaugeMetricValue(rtm.GCCPUFraction))
	ms.PutGaugeMetric("GCSys", GaugeMetricValue(rtm.GCSys))
	ms.PutGaugeMetric("HeapAlloc", GaugeMetricValue(rtm.HeapAlloc))
	ms.PutGaugeMetric("HeapIdle", GaugeMetricValue(rtm.HeapIdle))
	ms.PutGaugeMetric("HeapInuse", GaugeMetricValue(rtm.HeapInuse))
	ms.PutGaugeMetric("HeapObjects", GaugeMetricValue(rtm.HeapObjects))
	ms.PutGaugeMetric("HeapReleased", GaugeMetricValue(rtm.HeapReleased))
	ms.PutGaugeMetric("HeapSys", GaugeMetricValue(rtm.HeapSys))
	ms.PutGaugeMetric("LastGC", GaugeMetricValue(rtm.LastGC))
	ms.PutGaugeMetric("Lookups", GaugeMetricValue(rtm.Lookups))
	ms.PutGaugeMetric("MCacheInuse", GaugeMetricValue(rtm.MCacheInuse))
	ms.PutGaugeMetric("MCacheSys", GaugeMetricValue(rtm.MCacheSys))
	ms.PutGaugeMetric("MSpanInuse", GaugeMetricValue(rtm.MSpanInuse))
	ms.PutGaugeMetric("MSpanSys", GaugeMetricValue(rtm.MSpanSys))
	ms.PutGaugeMetric("Mallocs", GaugeMetricValue(rtm.Mallocs))
	ms.PutGaugeMetric("NextGC", GaugeMetricValue(rtm.NextGC))
	ms.PutGaugeMetric("NumForcedGC", GaugeMetricValue(rtm.NumForcedGC))
	ms.PutGaugeMetric("NumGC", GaugeMetricValue(rtm.NumGC))
	ms.PutGaugeMetric("OtherSys", GaugeMetricValue(rtm.OtherSys))
	ms.PutGaugeMetric("PauseTotalNs", GaugeMetricValue(rtm.PauseTotalNs))
	ms.PutGaugeMetric("StackInuse", GaugeMetricValue(rtm.StackInuse))
	ms.PutGaugeMetric("StackSys", GaugeMetricValue(rtm.StackSys))
	ms.PutGaugeMetric("Sys", GaugeMetricValue(rtm.Sys))
	ms.PutGaugeMetric("TotalAlloc", GaugeMetricValue(rtm.TotalAlloc))
	ms.PutGaugeMetric("RandomValue", GaugeMetricValue(rand.Float64()))
	ms.PutCounterMetric("PollCount", CounterMetricValue(1))
}

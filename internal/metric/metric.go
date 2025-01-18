package metric

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
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
	collectMetric := func(metricType Metric, metricName string, metricValue any) {
		var err error
		switch metricType {
		case GaugeMetric:
			err = ms.PutGaugeMetric(metricName, metricValue.(GaugeMetricValue))
		case CounterMetric:
			err = ms.PutCounterMetric(metricName, metricValue.(CounterMetricValue))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: collecting %s metric %s:%v\n", metricType, metricName, err)
		}
	}

	collectMetric(GaugeMetric, "Alloc", GaugeMetricValue(rtm.Alloc))
	collectMetric(GaugeMetric, "BuckHashSys", GaugeMetricValue(rtm.BuckHashSys))
	collectMetric(GaugeMetric, "Frees", GaugeMetricValue(rtm.Frees))
	collectMetric(GaugeMetric, "GCCPUFraction", GaugeMetricValue(rtm.GCCPUFraction))
	collectMetric(GaugeMetric, "GCSys", GaugeMetricValue(rtm.GCSys))
	collectMetric(GaugeMetric, "HeapAlloc", GaugeMetricValue(rtm.HeapAlloc))
	collectMetric(GaugeMetric, "HeapIdle", GaugeMetricValue(rtm.HeapIdle))
	collectMetric(GaugeMetric, "HeapInuse", GaugeMetricValue(rtm.HeapInuse))
	collectMetric(GaugeMetric, "HeapObjects", GaugeMetricValue(rtm.HeapObjects))
	collectMetric(GaugeMetric, "HeapReleased", GaugeMetricValue(rtm.HeapReleased))
	collectMetric(GaugeMetric, "HeapSys", GaugeMetricValue(rtm.HeapSys))
	collectMetric(GaugeMetric, "LastGC", GaugeMetricValue(rtm.LastGC))
	collectMetric(GaugeMetric, "Lookups", GaugeMetricValue(rtm.Lookups))
	collectMetric(GaugeMetric, "MCacheInuse", GaugeMetricValue(rtm.MCacheInuse))
	collectMetric(GaugeMetric, "MCacheSys", GaugeMetricValue(rtm.MCacheSys))
	collectMetric(GaugeMetric, "MSpanInuse", GaugeMetricValue(rtm.MSpanInuse))
	collectMetric(GaugeMetric, "MSpanSys", GaugeMetricValue(rtm.MSpanSys))
	collectMetric(GaugeMetric, "Mallocs", GaugeMetricValue(rtm.Mallocs))
	collectMetric(GaugeMetric, "NextGC", GaugeMetricValue(rtm.NextGC))
	collectMetric(GaugeMetric, "NumForcedGC", GaugeMetricValue(rtm.NumForcedGC))
	collectMetric(GaugeMetric, "NumGC", GaugeMetricValue(rtm.NumGC))
	collectMetric(GaugeMetric, "OtherSys", GaugeMetricValue(rtm.OtherSys))
	collectMetric(GaugeMetric, "PauseTotalNs", GaugeMetricValue(rtm.PauseTotalNs))
	collectMetric(GaugeMetric, "StackInuse", GaugeMetricValue(rtm.StackInuse))
	collectMetric(GaugeMetric, "StackSys", GaugeMetricValue(rtm.StackSys))
	collectMetric(GaugeMetric, "Sys", GaugeMetricValue(rtm.Sys))
	collectMetric(GaugeMetric, "TotalAlloc", GaugeMetricValue(rtm.TotalAlloc))
	collectMetric(GaugeMetric, "RandomValue", GaugeMetricValue(rand.Float64()))

	collectMetric(CounterMetric, "PollCount", CounterMetricValue(1))
}

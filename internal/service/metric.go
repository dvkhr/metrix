package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"

	//----
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)

type GaugeMetricValue float64
type CounterMetricValue int64

type Metrics struct {
	ID    string              `json:"id"`              // имя метрики
	MType MetricType          `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *CounterMetricValue `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *GaugeMetricValue   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var ErrUninitializedStorage = errors.New("storage is not initialized")
var ErrInvalidMetricName = errors.New("invalid metric name")
var ErrUnkonownMetric = errors.New("unknown metric")

type MetricStorage interface {
	Save(ctx context.Context, mt Metrics) error
	List(ctx context.Context) (*map[string]Metrics, error)
}

func CollectMetrics(ctx context.Context, ms MetricStorage) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	collectMetric := func(metricType MetricType, metricName string, metricValue any) {
		var mt Metrics
		switch metricType {
		case GaugeMetric:
			temp := metricValue.(GaugeMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Value: &temp}
		case CounterMetric:
			temp := metricValue.(CounterMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Delta: &temp}
		}
		err := ms.Save(ctx, mt)
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

func DumpMetrics(ms MetricStorage, wr io.Writer) error {

	ctx := context.TODO()
	mtrx, err := ms.List(ctx)
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
func RestoreMetrics(ms MetricStorage, rd io.Reader) error {
	ctx := context.TODO()
	var data []byte

	data, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	stor, err := ms.List(ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, stor)

	return err
}

func CollectMetricsOS(ctx context.Context, metrics chan Metrics) {
	fmt.Println("Run CollectMetricsOS")
	collectMetric := func(metricType MetricType, metricName string, metricValue any) {
		var mt Metrics
		switch metricType {
		case GaugeMetric:
			temp := metricValue.(GaugeMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Value: &temp}
		case CounterMetric:
			temp := metricValue.(CounterMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Delta: &temp}
		}
		metrics <- mt
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "collecting memory metrics: %v\n", err)
	}

	collectMetric(GaugeMetric, "TotalMemory", GaugeMetricValue(vMem.Total))
	collectMetric(GaugeMetric, "FreeMemory", GaugeMetricValue(vMem.Free))

	CPUutilization, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collecting CPU metrics: %v\n", err)
	}
	for i, iCPUtil := range CPUutilization {
		collectMetric(GaugeMetric, fmt.Sprintf("CPUutilization%d", i+1), GaugeMetricValue(iCPUtil))
	}
}

func CollectMetricsCh(ctx context.Context, metrics chan Metrics) {
	fmt.Println("Run CollectMetricsCh")

	collectMetric := func(metricType MetricType, metricName string, metricValue any) {
		var mt Metrics
		switch metricType {
		case GaugeMetric:
			temp := metricValue.(GaugeMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Value: &temp}
		case CounterMetric:
			temp := metricValue.(CounterMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Delta: &temp}
		}
		metrics <- mt
	}

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

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

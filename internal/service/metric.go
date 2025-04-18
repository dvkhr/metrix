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
	"github.com/dvkhr/metrix.git/internal/logging"
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
	logging.Logg.Info("+++Run CollectMetricsOS+++\n")

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
		logging.Logg.Error("collecting memory metrics: %v\n", err)
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

	gaugeMetrics := []struct {
		Name  string
		Value GaugeMetricValue
	}{
		{"Alloc", GaugeMetricValue(rtm.Alloc)},
		{"BuckHashSys", GaugeMetricValue(rtm.BuckHashSys)},
		{"Frees", GaugeMetricValue(rtm.Frees)},
		{"GCCPUFraction", GaugeMetricValue(rtm.GCCPUFraction)},
		{"GCSys", GaugeMetricValue(rtm.GCSys)},
		{"HeapAlloc", GaugeMetricValue(rtm.HeapAlloc)},
		{"HeapIdle", GaugeMetricValue(rtm.HeapIdle)},
		{"HeapInuse", GaugeMetricValue(rtm.HeapInuse)},
		{"HeapObjects", GaugeMetricValue(rtm.HeapObjects)},
		{"HeapReleased", GaugeMetricValue(rtm.HeapReleased)},
		{"HeapSys", GaugeMetricValue(rtm.HeapSys)},
		{"LastGC", GaugeMetricValue(rtm.LastGC)},
		{"Lookups", GaugeMetricValue(rtm.Lookups)},
		{"MCacheInuse", GaugeMetricValue(rtm.MCacheInuse)},
		{"MCacheSys", GaugeMetricValue(rtm.MCacheSys)},
		{"MSpanInuse", GaugeMetricValue(rtm.MSpanInuse)},
		{"MSpanSys", GaugeMetricValue(rtm.MSpanSys)},
		{"Mallocs", GaugeMetricValue(rtm.Mallocs)},
		{"NextGC", GaugeMetricValue(rtm.NextGC)},
		{"NumForcedGC", GaugeMetricValue(rtm.NumForcedGC)},
		{"NumGC", GaugeMetricValue(rtm.NumGC)},
		{"OtherSys", GaugeMetricValue(rtm.OtherSys)},
		{"PauseTotalNs", GaugeMetricValue(rtm.PauseTotalNs)},
		{"StackInuse", GaugeMetricValue(rtm.StackInuse)},
		{"StackSys", GaugeMetricValue(rtm.StackSys)},
		{"Sys", GaugeMetricValue(rtm.Sys)},
		{"TotalAlloc", GaugeMetricValue(rtm.TotalAlloc)},
		{"RandomValue", GaugeMetricValue(rand.Float64())},
	}
	for _, gm := range gaugeMetrics {
		collectMetric(GaugeMetric, gm.Name, gm.Value)
	}

	collectMetric(CounterMetric, "PollCount", CounterMetricValue(1))
}

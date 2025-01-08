package metric

import (
	"errors"
	"math/rand"
	"runtime"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (m *MemStorage) NewMemStorage() {
	m.gauge = make(map[string]float64)
	m.counter = make(map[string]int64)
}

func (m *MemStorage) PutGaugeMetric(name string, value float64) {
	m.gauge[name] = value
}

func (m *MemStorage) PutCounterMetric(name string, value int64) {
	m.counter[name] += value
}

func (m *MemStorage) GetGaugeMetric(name string) (float64, error) {
	if val, ok := m.gauge[name]; ok {
		return val, nil
	} else {
		return 0.0, errors.New("no such metric")
	}
}

func (m *MemStorage) GetCounterMetric(name string) (int64, error) {
	if val, ok := m.counter[name]; ok {
		return val, nil
	} else {
		return 0, errors.New("no such metric")
	}
}
func (m *MemStorage) CollectMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.PutGaugeMetric("Alloc", float64(rtm.Alloc))
	m.PutGaugeMetric("BuckHashSys", float64(rtm.BuckHashSys))
	m.PutGaugeMetric("Frees", float64(rtm.Frees))
	m.PutGaugeMetric("GCCPUFraction", float64(rtm.GCCPUFraction))
	m.PutGaugeMetric("GCSys", float64(rtm.GCSys))
	m.PutGaugeMetric("HeapAlloc", float64(rtm.HeapAlloc))
	m.PutGaugeMetric("HeapIdle", float64(rtm.HeapIdle))
	m.PutGaugeMetric("HeapInuse", float64(rtm.HeapInuse))
	m.PutGaugeMetric("HeapObjects", float64(rtm.HeapObjects))
	m.PutGaugeMetric("HeapReleased", float64(rtm.HeapReleased))
	m.PutGaugeMetric("HeapSys", float64(rtm.HeapSys))
	m.PutGaugeMetric("LastGC", float64(rtm.LastGC))
	m.PutGaugeMetric("Lookups", float64(rtm.Lookups))
	m.PutGaugeMetric("MCacheInuse", float64(rtm.MCacheInuse))
	m.PutGaugeMetric("MCacheSys", float64(rtm.MCacheSys))
	m.PutGaugeMetric("MSpanInuse", float64(rtm.MSpanInuse))
	m.PutGaugeMetric("MSpanSys", float64(rtm.MSpanSys))
	m.PutGaugeMetric("Mallocs", float64(rtm.Mallocs))
	m.PutGaugeMetric("NextGC", float64(rtm.NextGC))
	m.PutGaugeMetric("NumForcedGC", float64(rtm.NumForcedGC))
	m.PutGaugeMetric("NumGC", float64(rtm.NumGC))
	m.PutGaugeMetric("OtherSys", float64(rtm.OtherSys))
	m.PutGaugeMetric("PauseTotalNs", float64(rtm.PauseTotalNs))
	m.PutGaugeMetric("StackInuse", float64(rtm.StackInuse))
	m.PutGaugeMetric("StackSys", float64(rtm.StackSys))
	m.PutGaugeMetric("Sys", float64(rtm.Sys))
	m.PutGaugeMetric("TotalAlloc", float64(rtm.TotalAlloc))
	m.PutGaugeMetric("RandomValue", rand.Float64())
	m.PutCounterMetric("PollCount", 1)
}

func (m *MemStorage) AllGaugeMetrics() map[string]float64 {
	return m.gauge
}

func (m *MemStorage) AllCounterMetrics() map[string]int64 {
	return m.counter
}

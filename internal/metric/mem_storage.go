package metric

import "errors"

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

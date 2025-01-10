package storage

import (
	"errors"
	"fmt"
)

type MetricType string

const (
	GaugeMetric   = MetricType("gauge")
	CounterMetric = MetricType("counter")
)

var ErrUninitializedStorage = errors.New("storage is not initialized")
var ErrInvalidMetricName = errors.New("invalid metric name")
var ErrUnkonownMetric = errors.New("unknown metric")

type GuageMetricValue float64
type CounterMetricValue int64

type MemStorage struct {
	data map[string]interface{}
}

func (ms *MemStorage) NewMemStorage() {
	ms.data = make(map[string]interface{})
}

func (ms *MemStorage) PutGaugeMetric(metricName string, metricValue GuageMetricValue) error {
	if ms.data == nil {
		return ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return ErrInvalidMetricName
	}
	ms.data[metricName] = metricValue

	return nil
}

func (ms *MemStorage) GetGaugeMetric(metricName string) (GuageMetricValue, error) {
	if ms.data == nil {
		return 0.0, ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return 0.0, ErrInvalidMetricName
	}
	if metricValue, ok := ms.data[metricName].(GuageMetricValue); ok {
		return metricValue, nil
	}
	return 0.0, ErrUnkonownMetric
}

func (ms *MemStorage) PutCounterMetric(metricName string, metricValue CounterMetricValue) error {
	if ms.data == nil {
		return ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return ErrInvalidMetricName
	}

	currentValue := ms.data[metricName]
	if currentValue == nil {
		ms.data[metricName] = metricValue
	} else {
		ms.data[metricName] = ms.data[metricName].(CounterMetricValue) + metricValue
	}

	return nil
}

func (ms *MemStorage) GetCounterMetric(metricName string) (CounterMetricValue, error) {
	if ms.data == nil {
		return 0, ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return 0, ErrInvalidMetricName
	}
	if metricValue, ok := ms.data[metricName].(CounterMetricValue); ok {
		return metricValue, nil
	}
	return 0, ErrUnkonownMetric
}

func (ms *MemStorage) MetricStrings() ([]string, error) {
	if ms.data == nil {
		return nil, ErrUninitializedStorage
	}

	metricStrings := make([]string, 0, len(ms.data))

	for metricName, metricValue := range ms.data {
		switch val := metricValue.(type) {
		case GuageMetricValue:
			metricStrings = append(metricStrings, fmt.Sprintf("%s/%s/%v",
				GaugeMetric, metricName, val))
		case CounterMetricValue:
			metricStrings = append(metricStrings, fmt.Sprintf("%s/%s/%v",
				CounterMetric, metricName, val))
		}
	}

	return metricStrings, nil
}

/*
func (m *MemStorage) AllGaugeMetrics() map[string]float64 {
	return m.gauge
}

func (m *MemStorage) AllCounterMetrics() map[string]int64 {
	return m.counter
}
func (m *MemStorage) ResetCounterMetrics() {
	clear(m.counter)
}
*/

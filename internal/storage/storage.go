package storage

import (
	"fmt"

	"github.com/dvkhr/metrix.git/internal/metric"
)

type MemStorage struct {
	data map[string]interface{}
}

func (ms *MemStorage) NewMemStorage() {
	ms.data = make(map[string]interface{})
}

func (ms *MemStorage) PutGaugeMetric(metricName string, metricValue metric.GaugeMetricValue) error {
	if ms.data == nil {
		return metric.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return metric.ErrInvalidMetricName
	}
	ms.data[metricName] = metricValue

	return nil
}

func (ms *MemStorage) GetGaugeMetric(metricName string) (metric.GaugeMetricValue, error) {
	if ms.data == nil {
		return 0.0, metric.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return 0.0, metric.ErrInvalidMetricName
	}
	if metricValue, ok := ms.data[metricName].(metric.GaugeMetricValue); ok {
		return metricValue, nil
	}
	return 0.0, metric.ErrUnkonownMetric
}

func (ms *MemStorage) PutCounterMetric(metricName string, metricValue metric.CounterMetricValue) error {
	if ms.data == nil {
		return metric.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return metric.ErrInvalidMetricName
	}

	currentValue := ms.data[metricName]
	if currentValue == nil {
		ms.data[metricName] = metricValue
	} else {
		ms.data[metricName] = ms.data[metricName].(metric.CounterMetricValue) + metricValue
	}

	return nil
}

func (ms *MemStorage) GetCounterMetric(metricName string) (metric.CounterMetricValue, error) {
	if ms.data == nil {
		return 0, metric.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return 0, metric.ErrInvalidMetricName
	}
	if metricValue, ok := ms.data[metricName].(metric.CounterMetricValue); ok {
		return metricValue, nil
	}
	return 0, metric.ErrUnkonownMetric
}

func (ms *MemStorage) MetricStrings() ([]string, error) {
	if ms.data == nil {
		return nil, metric.ErrUninitializedStorage
	}

	metricStrings := make([]string, 0, len(ms.data))

	for metricName, metricValue := range ms.data {
		switch val := metricValue.(type) {
		case metric.GaugeMetricValue:
			metricStrings = append(metricStrings, fmt.Sprintf("%s/%s/%v",
				metric.GaugeMetric, metricName, val))
		case metric.CounterMetricValue:
			metricStrings = append(metricStrings, fmt.Sprintf("%s/%s/%v",
				metric.CounterMetric, metricName, val))
		}
	}

	return metricStrings, nil
}

func (ms *MemStorage) AllMetrics() (*map[string]interface{}, error) {
	if ms.data == nil {
		return nil, metric.ErrUninitializedStorage
	}
	return &ms.data, nil
}

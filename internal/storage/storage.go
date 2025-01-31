package storage

import (
	"github.com/dvkhr/metrix.git/internal/metric"
)

type MemStorage struct {
	data map[string]metric.Metrics
}

func (ms *MemStorage) NewMemStorage() {
	ms.data = make(map[string]metric.Metrics)
}

func (ms *MemStorage) Save(mt metric.Metrics) error {
	if ms.data == nil {
		return metric.ErrUninitializedStorage
	}
	if len(mt.ID) == 0 {
		return metric.ErrInvalidMetricName
	}
	if mt.MType == metric.GaugeMetric {
		ms.data[mt.ID] = mt
	} else if mt.MType == metric.CounterMetric {
		if ms.data[mt.ID].Delta != nil {
			*ms.data[mt.ID].Delta += *mt.Delta
		} else {
			ms.data[mt.ID] = mt
		}

	} else {
		return metric.ErrInvalidMetricName
	}
	return nil
}

func (ms *MemStorage) Get(metricName string) (*metric.Metrics, error) {
	if ms.data == nil {
		return nil, metric.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return nil, metric.ErrInvalidMetricName
	}
	if m, ok := ms.data[metricName]; ok {
		return &m, nil
	}
	return nil, metric.ErrUnkonownMetric
}

func (ms *MemStorage) List() (*map[string]metric.Metrics, error) {
	if ms.data == nil {
		return nil, metric.ErrUninitializedStorage
	}

	return &ms.data, nil
}

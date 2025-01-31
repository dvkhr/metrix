package storage

import (
	"github.com/dvkhr/metrix.git/internal/service"
)

type MemStorage struct {
	data map[string]service.Metrics
}

func (ms *MemStorage) NewMemStorage() {
	ms.data = make(map[string]service.Metrics)
}

func (ms *MemStorage) Save(mt service.Metrics) error {
	if ms.data == nil {
		return service.ErrUninitializedStorage
	}
	if len(mt.ID) == 0 {
		return service.ErrInvalidMetricName
	}
	if mt.MType == service.GaugeMetric {
		ms.data[mt.ID] = mt
	} else if mt.MType == service.CounterMetric {
		if ms.data[mt.ID].Delta != nil {
			*ms.data[mt.ID].Delta += *mt.Delta
		} else {
			ms.data[mt.ID] = mt
		}

	} else {
		return service.ErrInvalidMetricName
	}
	return nil
}

func (ms *MemStorage) Get(metricName string) (*service.Metrics, error) {
	if ms.data == nil {
		return nil, service.ErrUninitializedStorage
	}
	if len(metricName) == 0 {
		return nil, service.ErrInvalidMetricName
	}
	if m, ok := ms.data[metricName]; ok {
		return &m, nil
	}
	return nil, service.ErrUnkonownMetric
}

func (ms *MemStorage) List() (*map[string]service.Metrics, error) {
	if ms.data == nil {
		return nil, service.ErrUninitializedStorage
	}

	return &ms.data, nil
}

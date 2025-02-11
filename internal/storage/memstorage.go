package storage

import (
	"context"

	"github.com/dvkhr/metrix.git/internal/service"
)

type MemStorage struct {
	data map[string]service.Metrics
}

func (ms *MemStorage) NewStorage() error {
	ms.data = make(map[string]service.Metrics)
	return nil
}

func (ms *MemStorage) Save(ctx context.Context, mt service.Metrics) error {
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

func (ms *MemStorage) Get(ctx context.Context, metricName string) (*service.Metrics, error) {
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

func (ms *MemStorage) List(ctx context.Context) (*map[string]service.Metrics, error) {
	if ms.data == nil {
		return nil, service.ErrUninitializedStorage
	}

	return &ms.data, nil
}

func (ms *MemStorage) ListSlice(ctx context.Context) ([]service.Metrics, error) {
	if ms.data == nil {
		return nil, service.ErrUninitializedStorage
	}
	metricsSlice := make([]service.Metrics, 0, len(ms.data))
	for _, metric := range ms.data {
		metricsSlice = append(metricsSlice, metric)
	}

	return metricsSlice, nil
}

func (ms *MemStorage) FreeStorage() error {
	return nil
}

func (ms *MemStorage) CheckStorage() error {
	if ms.data == nil {
		return service.ErrUninitializedStorage
	}
	return nil
}

func (ms *MemStorage) SaveAll(ctx context.Context, mt *[]service.Metrics) error {
	if ms.data == nil {
		return service.ErrUninitializedStorage
	}
	if len(*mt) == 0 {
		return service.ErrInvalidMetricName
	}
	for _, metric := range *mt {
		if metric.MType == service.GaugeMetric {
			ms.data[metric.ID] = metric
		} else if metric.MType == service.CounterMetric {
			if ms.data[metric.ID].Delta != nil {
				*ms.data[metric.ID].Delta += *metric.Delta
			} else {
				ms.data[metric.ID] = metric
			}
		} else {
			return service.ErrInvalidMetricName
		}
	}

	return nil
}

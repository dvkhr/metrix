package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/dvkhr/metrix.git/internal/service"
)

type FileStorage struct {
	FileStoragePath string
	file            *os.File
}

func (ms *FileStorage) NewStorage() error {
	var err error
	ms.file, err = os.OpenFile(ms.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	return err
}

func (ms *FileStorage) Save(ctx context.Context, mt service.Metrics) error {
	defer ms.file.Sync()

	if ms.file == nil {
		return service.ErrUninitializedStorage
	}

	if len(mt.ID) == 0 {
		return service.ErrInvalidMetricName
	}

	mtrx, err := ms.List(ctx)
	if err != nil {
		return err
	}

	if mt.MType == service.GaugeMetric {
		(*mtrx)[mt.ID] = mt
	} else if mt.MType == service.CounterMetric {
		if (*mtrx)[mt.ID].Delta != nil {
			*(*mtrx)[mt.ID].Delta += *mt.Delta
		} else {
			(*mtrx)[mt.ID] = mt
		}
	} else {
		return service.ErrInvalidMetricName
	}

	data, err := json.MarshalIndent(mtrx, "", "  ")
	if err != nil {
		return err
	}

	err = ms.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = ms.file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = ms.file.Write(data)
	return err
}

func (ms *FileStorage) SaveAll(ctx context.Context, mt *[]service.Metrics) error {
	defer ms.file.Sync()

	if ms.file == nil {
		return service.ErrUninitializedStorage
	}

	if len(*mt) == 0 {
		return service.ErrInvalidMetricName
	}

	mtrx, err := ms.List(ctx)
	if err != nil {
		return err
	}
	for _, metric := range *mt {
		if metric.MType == service.GaugeMetric {
			(*mtrx)[metric.ID] = metric
		} else if metric.MType == service.CounterMetric {
			if (*mtrx)[metric.ID].Delta != nil {
				*(*mtrx)[metric.ID].Delta += *metric.Delta
			} else {
				(*mtrx)[metric.ID] = metric
			}
		} else {
			return service.ErrInvalidMetricName
		}

	}

	data, err := json.MarshalIndent(mtrx, "", "  ")
	if err != nil {
		return err
	}

	err = ms.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = ms.file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = ms.file.Write(data)
	return err
}

func (ms *FileStorage) Get(ctx context.Context, metricName string) (*service.Metrics, error) {
	if ms.file == nil {
		return nil, service.ErrUninitializedStorage
	}

	if len(metricName) == 0 {
		return nil, service.ErrInvalidMetricName
	}

	mtrx, err := ms.List(ctx)
	if err != nil {
		return nil, err
	}

	if m, ok := (*mtrx)[metricName]; ok {
		return &m, nil
	}
	return nil, service.ErrUnknownMetric
}

func (ms *FileStorage) List(ctx context.Context) (*map[string]service.Metrics, error) {
	if ms.file == nil {
		return nil, service.ErrUninitializedStorage
	}
	var data []byte
	var mtrx map[string]service.Metrics

	ms.file.Seek(0, 0)

	data, err := io.ReadAll(ms.file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		mtrx = make(map[string]service.Metrics)
		return &mtrx, nil
	}

	err = json.Unmarshal(data, &mtrx)
	if err != nil {
		return nil, err
	}
	return &mtrx, nil
}

func (ms *FileStorage) FreeStorage() error {
	return ms.file.Close()
}

func (ms *FileStorage) CheckStorage() error {
	if ms.file == nil {
		return service.ErrUninitializedStorage
	}
	return nil
}
func (ms *FileStorage) ListSlice(ctx context.Context) ([]service.Metrics, error) {
	if ms.file == nil {
		return nil, service.ErrUninitializedStorage
	}
	var data []byte
	var mtrx []service.Metrics

	ms.file.Seek(0, 0)

	data, err := io.ReadAll(ms.file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		mtrx = make([]service.Metrics, 0, len(data))
		return mtrx, nil
	}

	err = json.Unmarshal(data, &mtrx)
	if err != nil {
		return nil, err
	}
	return mtrx, nil
}

package storage

import (
	"context"
	"testing"

	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_NewMemStorage(t *testing.T) {
	type fields struct {
		data map[string]service.Metrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Successful",
			fields: fields{
				data: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				data: tt.fields.data,
			}

			ms.NewStorage()

			if ms.data == nil {
				t.Errorf("MemStorage.NewMemStroage() must initialize MemStroage.data map")
			}
		})

	}
}

func TestMemStorageSave(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{} // data == nil
		err := storage.Save(ctx, service.Metrics{ID: "test_metric"})
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})

	t.Run("Error: Invalid metric name", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()
		err := storage.Save(ctx, service.Metrics{ID: "", MType: service.GaugeMetric})
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})

	t.Run("Save gauge metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		value := service.GaugeMetricValue(42.0)
		metric := service.Metrics{
			ID:    "test_gauge",
			MType: service.GaugeMetric,
			Value: &value,
		}

		err := storage.Save(ctx, metric)
		assert.NoError(t, err)

		savedMetric, exists := storage.data["test_gauge"]
		assert.True(t, exists)
		assert.Equal(t, *savedMetric.Value, value)
	})

	t.Run("Save counter metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		delta1 := service.CounterMetricValue(10)
		delta2 := service.CounterMetricValue(20)
		result := service.CounterMetricValue(30)

		metric1 := service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &delta1,
		}
		err := storage.Save(ctx, metric1)
		assert.NoError(t, err)

		savedMetric, exists := storage.data["test_counter"]
		assert.True(t, exists)
		assert.Equal(t, *savedMetric.Delta, delta1)

		metric2 := service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &delta2,
		}
		err = storage.Save(ctx, metric2)
		assert.NoError(t, err)

		savedMetric, exists = storage.data["test_counter"]
		assert.True(t, exists)
		assert.Equal(t, *savedMetric.Delta, result)
	})
}

func TestMemStorageGet(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{}
		metric, err := storage.Get(ctx, "test_metric")
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
		assert.Nil(t, metric)
	})

	t.Run("Error: Invalid metric name", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()
		metric, err := storage.Get(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
		assert.Nil(t, metric)
	})

	t.Run("Success: Get gauge metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		value := service.GaugeMetricValue(42.0)
		metric := service.Metrics{
			ID:    "test_gauge",
			MType: service.GaugeMetric,
			Value: &value,
		}
		storage.data["test_gauge"] = metric

		result, err := storage.Get(ctx, "test_gauge")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test_gauge", result.ID)
		assert.Equal(t, service.GaugeMetric, result.MType)
		assert.Equal(t, service.GaugeMetricValue(42.0), *result.Value)
	})

	t.Run("Success: Get counter metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		delta := service.CounterMetricValue(10)
		metric := service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &delta,
		}
		storage.data["test_counter"] = metric

		result, err := storage.Get(ctx, "test_counter")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test_counter", result.ID)
		assert.Equal(t, service.CounterMetric, result.MType)
		assert.Equal(t, service.CounterMetricValue(10), *result.Delta)
	})

	t.Run("Error: Unknown metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		metric, err := storage.Get(ctx, "unknown_metric")
		assert.Error(t, err)
		assert.Equal(t, service.ErrUnknownMetric, err)
		assert.Nil(t, metric)
	})
}

func TestMemStorageList(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{} // data == nil
		result, err := storage.List(ctx)
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
		assert.Nil(t, result)
	})

	t.Run("Success: Empty storage", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		result, err := storage.List(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, *result)
	})

	t.Run("Success: Non-empty storage", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		gaugeValue := service.GaugeMetricValue(42.0)
		counterDelta := service.CounterMetricValue(10)

		storage.data["test_gauge"] = service.Metrics{
			ID:    "test_gauge",
			MType: service.GaugeMetric,
			Value: &gaugeValue,
		}
		storage.data["test_counter"] = service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &counterDelta,
		}

		result, err := storage.List(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, 2, len(*result))

		assert.Equal(t, gaugeValue, *(*result)["test_gauge"].Value)
		assert.Equal(t, counterDelta, *(*result)["test_counter"].Delta)
	})
}

func TestMemStorageListSlice(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{}
		result, err := storage.ListSlice(ctx)
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
		assert.Nil(t, result)
	})

	t.Run("Success: Empty storage", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		result, err := storage.ListSlice(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("Success: Non-empty storage", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		gaugeValue := service.GaugeMetricValue(42.0)
		counterDelta := service.CounterMetricValue(10)

		storage.data["test_gauge"] = service.Metrics{
			ID:    "test_gauge",
			MType: service.GaugeMetric,
			Value: &gaugeValue,
		}
		storage.data["test_counter"] = service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &counterDelta,
		}

		result, err := storage.ListSlice(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, 2, len(result))

		foundGauge := false
		foundCounter := false

		for _, metric := range result {
			if metric.ID == "test_gauge" {
				assert.Equal(t, service.GaugeMetric, metric.MType)
				assert.Equal(t, gaugeValue, *metric.Value)
				foundGauge = true
			} else if metric.ID == "test_counter" {
				assert.Equal(t, service.CounterMetric, metric.MType)
				assert.Equal(t, counterDelta, *metric.Delta)
				foundCounter = true
			}
		}

		assert.True(t, foundGauge)
		assert.True(t, foundCounter)
	})
}

func TestCheckStorage(t *testing.T) {
	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{}

		err := storage.CheckStorage()
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})

	t.Run("Success: Initialized storage", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		err := storage.CheckStorage()
		assert.NoError(t, err)
	})
}

func TestMemStorageSaveAll(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &MemStorage{}
		metrics := []service.Metrics{}
		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})

	t.Run("Error: Empty metrics slice", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		metrics := []service.Metrics{}
		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})

	t.Run("Success: Save multiple metrics", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		gaugeValue1 := service.GaugeMetricValue(42.0)
		gaugeValue2 := service.GaugeMetricValue(100.0)
		counterDelta1 := service.CounterMetricValue(10)
		counterDelta2 := service.CounterMetricValue(20)

		metrics := []service.Metrics{
			{ID: "gauge1", MType: service.GaugeMetric, Value: &gaugeValue1},
			{ID: "gauge2", MType: service.GaugeMetric, Value: &gaugeValue2},
			{ID: "counter1", MType: service.CounterMetric, Delta: &counterDelta1},
			{ID: "counter2", MType: service.CounterMetric, Delta: &counterDelta2},
		}

		err := storage.SaveAll(ctx, &metrics)
		assert.NoError(t, err)

		assert.Equal(t, 4, len(storage.data))

		assert.Equal(t, gaugeValue1, *storage.data["gauge1"].Value)
		assert.Equal(t, gaugeValue2, *storage.data["gauge2"].Value)
		assert.Equal(t, counterDelta1, *storage.data["counter1"].Delta)
		assert.Equal(t, counterDelta2, *storage.data["counter2"].Delta)
	})

	t.Run("Success: Update counter metric", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		initialDelta := service.CounterMetricValue(10)
		storage.data["counter1"] = service.Metrics{
			ID:    "counter1",
			MType: service.CounterMetric,
			Delta: &initialDelta,
		}

		newDelta := service.CounterMetricValue(20)
		metrics := []service.Metrics{
			{ID: "counter1", MType: service.CounterMetric, Delta: &newDelta},
		}

		err := storage.SaveAll(ctx, &metrics)
		assert.NoError(t, err)

		assert.Equal(t, service.CounterMetricValue(30), *storage.data["counter1"].Delta)
	})

	t.Run("Error: Invalid metric type", func(t *testing.T) {
		storage := &MemStorage{}
		_ = storage.NewStorage()

		metrics := []service.Metrics{
			{ID: "invalid_metric", MType: "unknown"},
		}

		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})
}

package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/stretchr/testify/assert"
)

func createTempFile(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_file.txt")

	cleanup := func() {
		os.Remove(tempFile)
	}

	return tempFile, cleanup
}

func TestFileNewStorage(t *testing.T) {
	t.Run("Success: Create and open file", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		err := storage.NewStorage()

		assert.NoError(t, err)

		assert.NotNil(t, storage.file)

		_, err = os.Stat(filePath)
		assert.NoError(t, err)
	})

	t.Run("Error: Invalid file path", func(t *testing.T) {
		invalidPath := "/invalid/path/to/file.txt"

		storage := &FileStorage{FileStoragePath: invalidPath}
		err := storage.NewStorage()

		assert.Error(t, err)

		assert.Nil(t, storage.file)
	})
}

func TestFileSave(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &FileStorage{}
		err := storage.Save(ctx, service.Metrics{})
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})

	t.Run("Error: Invalid metric name", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		err := storage.Save(ctx, service.Metrics{})
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})

	t.Run("Success: Save gauge metric", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		gaugeValue := service.GaugeMetricValue(42.0)
		metric := service.Metrics{
			ID:    "test_gauge",
			MType: service.GaugeMetric,
			Value: &gaugeValue,
		}

		err := storage.Save(ctx, metric)
		assert.NoError(t, err)

		data, err := os.ReadFile(filePath)
		assert.NoError(t, err)

		var savedData map[string]service.Metrics
		err = json.Unmarshal(data, &savedData)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(savedData))
		assert.Equal(t, gaugeValue, *savedData["test_gauge"].Value)
	})

	t.Run("Success: Save counter metric", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		delta1 := service.CounterMetricValue(10)
		delta2 := service.CounterMetricValue(20)
		result := service.CounterMetricValue(30)

		metric1 := service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &delta1,
		}
		metric2 := service.Metrics{
			ID:    "test_counter",
			MType: service.CounterMetric,
			Delta: &delta2,
		}

		err := storage.Save(ctx, metric1)
		assert.NoError(t, err)

		err = storage.Save(ctx, metric2)
		assert.NoError(t, err)

		data, err := os.ReadFile(filePath)
		assert.NoError(t, err)

		var savedData map[string]service.Metrics
		err = json.Unmarshal(data, &savedData)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(savedData))
		assert.Equal(t, result, *savedData["test_counter"].Delta)
	})

	t.Run("Error: Invalid metric type", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		err := storage.Save(ctx, service.Metrics{ID: "test_invalid", MType: "unknown"})
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})
}

func TestFileSaveAll(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &FileStorage{}
		metrics := []service.Metrics{}
		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})

	t.Run("Error: Empty metrics slice", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		metrics := []service.Metrics{}
		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})

	t.Run("Success: Save multiple metrics", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
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

		data, err := os.ReadFile(filePath)
		assert.NoError(t, err)

		var savedData map[string]service.Metrics
		err = json.Unmarshal(data, &savedData)
		assert.NoError(t, err)

		assert.Equal(t, 4, len(savedData))
		assert.Equal(t, gaugeValue1, *savedData["gauge1"].Value)
		assert.Equal(t, gaugeValue2, *savedData["gauge2"].Value)
		assert.Equal(t, counterDelta1, *savedData["counter1"].Delta)
		assert.Equal(t, counterDelta2, *savedData["counter2"].Delta)
	})

	t.Run("Success: Update counter metric", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		counterDelta1 := service.CounterMetricValue(10)
		counterDelta2 := service.CounterMetricValue(20)
		result := service.CounterMetricValue(30)

		metrics := []service.Metrics{
			{ID: "counter1", MType: service.CounterMetric, Delta: &counterDelta1},
		}
		err := storage.SaveAll(ctx, &metrics)
		assert.NoError(t, err)

		metrics = []service.Metrics{
			{ID: "counter1", MType: service.CounterMetric, Delta: &counterDelta2},
		}
		err = storage.SaveAll(ctx, &metrics)
		assert.NoError(t, err)

		data, err := os.ReadFile(filePath)
		assert.NoError(t, err)

		var savedData map[string]service.Metrics
		err = json.Unmarshal(data, &savedData)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(savedData))
		assert.Equal(t, result, *savedData["counter1"].Delta)
	})

	t.Run("Error: Invalid metric type", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		metrics := []service.Metrics{
			{ID: "invalid_metric", MType: "unknown"},
		}

		err := storage.SaveAll(ctx, &metrics)
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})
}

func TestFileFreeStorage(t *testing.T) {
	t.Run("Success: Close file", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		err := storage.NewStorage()
		assert.NoError(t, err)

		err = storage.FreeStorage()
		assert.NoError(t, err)

		assert.NotNil(t, storage.file)
	})

	t.Run("Error: Close uninitialized file", func(t *testing.T) {
		storage := &FileStorage{}

		err := storage.FreeStorage()
		assert.Error(t, err)
	})
}

func TestFileCheckStorage(t *testing.T) {
	t.Run("Success: Initialized storage", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		err := storage.NewStorage()
		assert.NoError(t, err)

		err = storage.CheckStorage()
		assert.NoError(t, err)
	})

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &FileStorage{} // file == nil

		err := storage.CheckStorage()
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
	})
}

func TestFileListSlice(t *testing.T) {
	ctx := context.Background()

	t.Run("Error: Uninitialized storage", func(t *testing.T) {
		storage := &FileStorage{}

		result, err := storage.ListSlice(ctx)
		assert.Error(t, err)
		assert.Equal(t, service.ErrUninitializedStorage, err)
		assert.Nil(t, result)
	})

	t.Run("Success: Empty file", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		result, err := storage.ListSlice(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("Success: Non-empty file", func(t *testing.T) {
		gaugeValue := service.GaugeMetricValue(42.0)
		counterDelta := service.CounterMetricValue(10)

		metrics := []service.Metrics{
			{ID: "gauge1", MType: service.GaugeMetric, Value: &gaugeValue},
			{ID: "counter1", MType: service.CounterMetric, Delta: &counterDelta},
		}

		content, err := json.Marshal(metrics)
		assert.NoError(t, err)

		filePath, cleanup := createTempFile(t)
		defer cleanup()

		err = os.WriteFile(filePath, content, 0644)
		assert.NoError(t, err)

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		result, err := storage.ListSlice(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))

		assert.Equal(t, "gauge1", result[0].ID)
		assert.Equal(t, service.GaugeMetric, result[0].MType)
		assert.Equal(t, gaugeValue, *result[0].Value)

		assert.Equal(t, "counter1", result[1].ID)
		assert.Equal(t, service.CounterMetric, result[1].MType)
		assert.Equal(t, counterDelta, *result[1].Delta)
	})

	t.Run("Error: Invalid JSON", func(t *testing.T) {
		filePath, cleanup := createTempFile(t)
		defer cleanup()

		err := os.WriteFile(filePath, []byte("{invalid_json}"), 0644)
		assert.NoError(t, err)

		storage := &FileStorage{FileStoragePath: filePath}
		_ = storage.NewStorage()

		result, err := storage.ListSlice(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

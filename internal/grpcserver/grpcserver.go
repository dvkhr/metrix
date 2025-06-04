package grpcserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/logging"

	"github.com/dvkhr/metrix.git/internal/service"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	MetricStorage service.MetricStorage
	SignKey       []byte
}

// BatchUpdate обрабатывает пакетную отправку метрик.
func (s *MetricsServer) BatchUpdate(ctx context.Context, req *pb.BatchRequest) (*pb.MetricResponse, error) {
	// Проверка подписи (если требуется)
	if len(s.SignKey) > 0 && req.Hash != "" {
		logging.Logg.Debug("Validating hash signature for the request")

		var jsonData []byte
		for _, metric := range req.Metrics {
			jsonData = append(jsonData, []byte(fmt.Sprintf(`{"id":"%s","type":%d,"value":%v,"delta":%v}`, metric.Id, metric.Type, metric.Value, metric.Delta))...)
		}
		jsonData = append(jsonData, s.SignKey...)
		hash := sha256.Sum256(jsonData)
		if hex.EncodeToString(hash[:]) != req.Hash {
			logging.Logg.Error("Invalid hash signature")
			return &pb.MetricResponse{
				Success: false,
				Message: "Invalid hash signature",
			}, nil
		}
		logging.Logg.Debug("Hash signature validation successful")
	}

	// Декомпрессия данных
	decompressedData, err := gzip.DecompressData(req.Data)
	if err != nil {
		logging.Logg.Error("Failed to decompress data: %v", err)
		return &pb.MetricResponse{
			Success: false,
			Message: "Failed to decompress data",
		}, nil
	}

	// Парсинг метрик из декомпрессированных данных
	var metrics []*pb.MetricRequest
	if err := json.Unmarshal(decompressedData, &metrics); err != nil {
		logging.Logg.Error("Failed to parse metrics from decompressed data: %v", err)
		return &pb.MetricResponse{
			Success: false,
			Message: "Failed to parse metrics",
		}, nil
	}

	// Сохранение метрик
	for _, metric := range metrics {
		mTemp := &service.Metrics{
			ID:    metric.Id,
			MType: service.MetricType(metric.Type),
		}
		if metric.Type == string(service.GaugeMetric) {
			value := service.GaugeMetricValue(metric.Value)
			mTemp.Value = &value
		} else if metric.Type == string(service.CounterMetric) {
			delta := service.CounterMetricValue(metric.Delta)
			mTemp.Delta = &delta
		}
		logging.Logg.Debug("Saving metric: ID=%s, Type=%d, Value=%v, Delta=%v",
			mTemp.ID, mTemp.MType, mTemp.Value, mTemp.Delta)

		if err := s.MetricStorage.Save(ctx, *mTemp); err != nil {
			return &pb.MetricResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to save metric %s: %v", metric.Id, err),
			}, nil
		}
	}

	return &pb.MetricResponse{
		Success: true,
		Message: "Batch update successful",
	}, nil
}

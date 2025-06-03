package grpcserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"

	"github.com/dvkhr/metrix.git/internal/service"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	MetricStorage service.MetricStorage
	SignKey       []byte
}

// BatchUpdate обрабатывает пакетную отправку метрик.
func (s *MetricsServer) BatchUpdate(ctx context.Context, req *pb.BatchRequest) (*pb.MetricResponse, error) {
	if len(s.SignKey) > 0 && req.Hash != "" {
		var jsonData []byte
		for _, metric := range req.Metrics {
			jsonData = append(jsonData, []byte(fmt.Sprintf(`{"id":"%s","type":%d,"value":%v,"delta":%v}`, metric.Id, metric.Type, metric.Value, metric.Delta))...)
		}
		jsonData = append(jsonData, s.SignKey...)
		hash := sha256.Sum256(jsonData)
		if hex.EncodeToString(hash[:]) != req.Hash {
			return &pb.MetricResponse{
				Success: false,
				Message: "Invalid hash signature",
			}, nil
		}
	}

	for _, metric := range req.Metrics {
		mTemp := &service.Metrics{
			ID:    metric.Id,
			MType: service.MetricType(metric.Type),
		}
		if metric.Type == pb.MetricType_GAUGE {
			value := service.GaugeMetricValue(metric.Value)
			mTemp.Value = &value
		} else if metric.Type == pb.MetricType_COUNTER {
			delta := service.CounterMetricValue(metric.Delta)
			mTemp.Delta = &delta
		}

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

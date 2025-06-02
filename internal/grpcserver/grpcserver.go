package grpcserver

import (
	"context"
	"fmt"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/service"
)

// MetricsServer представляет реализацию gRPC-сервера для работы с метриками.
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	MetricStorage service.MetricStorage
}

// SendMetric обрабатывает gRPC-запрос на отправку одной метрики.
//
// Метод принимает метрику в формате MetricRequest, проверяет её корректность,
// сохраняет в хранилище и возвращает результат операции в виде MetricResponse.
//
// Параметры:
// - ctx: Контекст для управления жизненным циклом запроса.
// - req: Запрос, содержащий данные метрики (идентификатор, тип, значение или дельту).
//
// Возвращаемые значения:
// - *pb.MetricResponse: Ответ, содержащий статус операции и сообщение.
// - error: Ошибка, если произошла проблема при обработке запроса.
func (s *MetricsServer) SendMetric(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {

	mTemp := &service.Metrics{
		ID:    req.Id,
		MType: service.MetricType(req.Type),
	}
	if req.Type == pb.MetricType_GAUGE {
		value := service.GaugeMetricValue(req.Value)
		mTemp.Value = &value
	} else if req.Type == pb.MetricType_COUNTER {
		delta := service.CounterMetricValue(req.Delta)
		mTemp.Delta = &delta
	}

	if err := s.MetricStorage.Save(ctx, *mTemp); err != nil {
		return &pb.MetricResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to save metric: %v", err),
		}, nil
	}

	return &pb.MetricResponse{
		Success: true,
		Message: "Metric saved successfully",
	}, nil
}

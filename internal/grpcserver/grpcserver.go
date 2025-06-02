package grpcserver

import (
	pb "github.com/dvkhr/metrix.git/internal/grpc"
	"github.com/dvkhr/metrix.git/internal/service"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	MetricStorage service.MetricStorage
}

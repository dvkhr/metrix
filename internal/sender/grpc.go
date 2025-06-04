package sender

import (
	"context"
	"fmt"
	"time"

	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SendStrategyGRPC struct {
	sender  string
	client  pb.MetricsServiceClient
	address string
}

func NewSendStrategyGRPC(address string) *SendStrategyGRPC {
	return &SendStrategyGRPC{sender: "gRPC", client: newGRPCClient(address), address: address}
}

// newGRPCClient создает и возвращает gRPC-клиент для взаимодействия с сервером.
func newGRPCClient(serverAddress string) pb.MetricsServiceClient {
	var conn *grpc.ClientConn
	var err error

	// Попытки подключения с задержкой между ними
	for i := 0; i < 3; i++ { // Максимум 3 попытки
		conn, err = grpc.NewClient(
			serverAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			break // Успешное подключение
		}
		logging.Logg.Error("gRPC connection attempt %d failed: %v", i+1, err)
		time.Sleep(1 * time.Second) // Пауза перед следующей попыткой
	}

	if err != nil {
		logging.Logg.Error("Failed to connect to gRPC server after retries: %v", err)
		return nil
	}

	return pb.NewMetricsServiceClient(conn)
}

func (ssg *SendStrategyGRPC) Send(ctx context.Context, compressedData []byte, signature string) error {
	batch := &pb.BatchRequest{
		Data: compressedData,
		Hash: signature,
	}
	message := fmt.Sprintf("Sending batch request with data length: %d, hash: %s", len(compressedData), signature)
	logging.Logg.Info(message)
	resp, err := ssg.client.BatchUpdate(ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to send batch via gRPC: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server returned error: %s", resp.Message)
	}

	return nil
}

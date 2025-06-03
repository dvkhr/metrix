// Package sender предоставляет функциональность для отправки метрик на удаленный сервер.
package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dvkhr/metrix.git/internal/crypto"
	pb "github.com/dvkhr/metrix.git/internal/grpc/proto"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/network"
	"github.com/dvkhr/metrix.git/internal/storage"
)

// SendOptions содержит параметры для отправки метрик.
type SendOptions struct {
	MemStorage    storage.MemStorage
	Client        *http.Client
	GRPCClient    pb.MetricsServiceClient
	ServerAddress string
	SignKey       []byte
	PublicKey     *rsa.PublicKey
	UseGRPC       bool
}

func SendMetrics(ctx context.Context, options SendOptions) error {

	logging.Logg.Info("+++Send metrics to server+++\n")

	allMetrics, err := options.MemStorage.ListSlice(ctx)
	if err == nil && len(allMetrics) > 0 {
		jsonMetric, err := json.Marshal(allMetrics)
		if err != nil {
			return err
		}

		var encryptedData string
		if options.PublicKey != nil {
			encryptedData, err = crypto.EncryptData(jsonMetric, options.PublicKey)
			if err != nil {
				logging.Logg.Error("Failed to encrypt data: %v", err)
				return err
			}
		} else {
			encryptedData = string(jsonMetric)
		}

		var requestBody bytes.Buffer

		gz := gzip.NewWriter(&requestBody)
		gz.Write([]byte(encryptedData))
		gz.Close()

		req, err := http.NewRequest("POST", buildAllMetricsURL(options.ServerAddress), &requestBody)
		if err != nil {
			fmt.Println(err)
		}

		if len(options.SignKey) > 0 {
			signBuf := jsonMetric
			signBuf = append(signBuf, ',')
			signBuf = append(signBuf, options.SignKey...)
			sign := sha256.Sum256(signBuf)
			req.Header.Set("HashSHA256", hex.EncodeToString(sign[:]))
		}
		clientIP, ipErr := network.GetOutboundIP()
		if ipErr != nil {
			logging.Logg.Warn("Failed to determine outbound IP: %v", ipErr)
		} else {
			req.Header.Set("X-Real-IP", clientIP)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Encoding", "gzip")
		resp, err := options.Client.Do(req)
		if err == nil {
			fmt.Println(resp.StatusCode)
			defer resp.Body.Close()
			var reader io.ReadCloser
			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					logging.Logg.Info("FAIL create gzip reader: %w", err)
				}
				defer reader.Close()
			default:
				reader = resp.Body
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				logging.Logg.Info("FAIL reader response body: %w", err)
				return err
			}
			fmt.Println(string(body))
		} else {
			return err
		}
		options.MemStorage.NewStorage()
	}
	return nil
}

func buildAllMetricsURL(serverAddress string) string {
	serverURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprint(serverAddress),
		Path:   "updates/",
	}
	return serverURL.String()
}

package sender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dvkhr/metrix.git/internal/gzip"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/network"
)

type SendStrategyHTTP struct {
	sender  string
	client  *http.Client
	address string
}

// newHTTPClient создает и возвращает HTTP-клиент с настройками таймаута и ограничениями для соединений.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}
}

func NewSendStrategyHTTP(address string) *SendStrategyHTTP {
	return &SendStrategyHTTP{sender: "HTTP", client: newHTTPClient(), address: address}
}

func (ssh *SendStrategyHTTP) Send(ctx context.Context, compressedData []byte, signature string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", buildAllMetricsURL(ssh.address), bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if signature != "" {
		req.Header.Set("HashSHA256", signature)
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

	resp, err := ssh.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	logging.Logg.Info("Server response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			logging.Logg.Error("Failed to read server response body: %v", readErr)
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		logging.Logg.Error("Server returned error: %s", string(body))
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewGzipReader(resp.Body)
		if err != nil {
			logging.Logg.Error("Failed to create gzip reader: %w", err)
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		logging.Logg.Error("Failed to read response body: %w", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logging.Logg.Debug("Server response body: %s", string(body))

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

// Package sender предоставляет функциональность для отправки метрик на удаленный сервер.
package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/storage"
)

func SendMetrics(ctx context.Context, mStor storage.MemStorage, cl *http.Client, serverAddress string, signKey []byte) error {

	logging.Logg.Info("+++Send metrics to server+++\n")

	allMetrics, err := mStor.ListSlice(ctx)
	if err == nil && len(allMetrics) > 0 {
		jsonMetric, err := json.Marshal(allMetrics)
		if err != nil {
			return err
		}

		var requestBody bytes.Buffer

		gz := gzip.NewWriter(&requestBody)
		gz.Write(jsonMetric)
		gz.Close()

		req, err := http.NewRequest("POST", buildAllMetricsURL(serverAddress), &requestBody)
		if err != nil {
			fmt.Println(err)
		}

		if len(signKey) > 0 {
			signBuf := jsonMetric
			signBuf = append(signBuf, ',')
			signBuf = append(signBuf, signKey...)
			sign := sha256.Sum256(signBuf)
			req.Header.Set("HashSHA256", hex.EncodeToString(sign[:]))
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Encoding", "gzip")
		resp, err := cl.Do(req)
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
		mStor.NewStorage()
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

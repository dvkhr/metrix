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

	"github.com/dvkhr/metrix.git/internal/storage"
)

func SendMetrics(mStor storage.MemStorage, ctx context.Context, cl *http.Client, serverAddress string, signKey []byte) error {

	fmt.Printf("+++Send metrics to server+++\n")
	allMetrics, err := mStor.ListSlice(ctx)
	if err == nil {
		jsonMetric, err := json.Marshal(allMetrics)
		if err != nil {
			return err
		}

		var requestBody bytes.Buffer

		gz := gzip.NewWriter(&requestBody)
		gz.Write(jsonMetric)
		gz.Close()

		req, err := http.NewRequest("POST", BuildAllMetricsURL(serverAddress), &requestBody)
		if err != nil {
			fmt.Println(err)
		}

		if len(signKey) > 0 {
			signBuf := requestBody.Bytes()
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
					fmt.Println("FAIL create gzip reader: %w", err)
				}
				defer reader.Close()
			default:
				reader = resp.Body
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				fmt.Println("FAIL reader response body: %w", err)
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

func BuildAllMetricsURL(serverAddress string) string {
	serverURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprint(serverAddress),
		Path:   "updates/",
	}
	return serverURL.String()
}

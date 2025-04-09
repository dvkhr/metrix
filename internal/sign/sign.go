package sign

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func SignCheck(h http.HandlerFunc, signKey []byte) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(signKey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		tempBuf, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(&tempBuf)

		agentSignStr := r.Header.Get("HashSHA256")
		if len(agentSignStr) > 0 {
			signatureValid := validateSignature(tempBuf.Bytes(), agentSignStr, signKey)
			if !signatureValid {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

// Читает тело запроса и сохраняет его в буфер
func readRequestBody(r *http.Request) (bytes.Buffer, error) {
	var tempBuf bytes.Buffer
	teeReader := io.TeeReader(r.Body, &tempBuf)
	_, err := io.ReadAll(teeReader)
	return tempBuf, err
}

// Проверяет подпись
func validateSignature(body []byte, agentSignStr string, signKey []byte) bool {
	serverSign := calculateServerSignature(body, signKey)

	agentSign, err := hex.DecodeString(agentSignStr)
	if err != nil || len(agentSign) != 32 {
		return false
	}

	return serverSign == [32]byte(agentSign)
}

// Вычисляет подпись сервера
func calculateServerSignature(body []byte, signKey []byte) [32]byte {
	signBuf := append(body, ',')
	signBuf = append(signBuf, signKey...)
	return sha256.Sum256(signBuf)
}

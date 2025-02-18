package sign_check

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func SignCheck(h http.HandlerFunc, signKey []byte) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(signKey) == 0 {
			return
		}

		agentSignStr := r.Header.Get("HashSHA256")

		if len(agentSignStr) == 0 {
			return //TODO make HTTP 500 error
		}

		signBuf, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		signBuf = append(signBuf, ',')
		signBuf = append(signBuf, signKey...)
		serverSign := sha256.Sum256(signBuf)

		agentSign, err := hex.DecodeString(agentSignStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if serverSign != [32]byte(agentSign) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	})
}

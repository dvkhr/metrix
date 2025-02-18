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
		var tempBuf bytes.Buffer
		if len(signKey) > 0 {
			if agentSignStr := r.Header.Get("HashSHA256"); len(agentSignStr) > 0 {

				teeReader := io.TeeReader(r.Body, &tempBuf)

				signBuf, err := io.ReadAll(teeReader)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

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
			}
		}

		r.Body = io.NopCloser(&tempBuf)
		h.ServeHTTP(w, r)
	})
}

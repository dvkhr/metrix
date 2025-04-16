package handlers

import (
	"encoding/json"
	"net/http"
)

func ReadAndUnmarshal(req *http.Request, v interface{}) error {

	const maxBodySize = 1 << 20 // 1 MB
	req.Body = http.MaxBytesReader(nil, req.Body, maxBodySize)

	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		return err
	}

	defer req.Body.Close()

	return nil
}

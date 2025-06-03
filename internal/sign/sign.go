// Package sign предоставляет инструменты для проверки подписи HTTP-запросов.
package sign

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// SignCheck создает middleware для проверки подписи HTTP-запроса.
// Middleware проверяет подпись запроса с использованием ключа (signKey).
// Если ключ пустой, проверка пропускается, и запрос передается дальше.
//
// Параметры:
// - h: Обработчик HTTP-запроса, который будет вызван после проверки подписи.
// - signKey: Ключ, используемый для проверки подписи.
//
// Возвращаемое значение:
// - http.HandlerFunc: Middleware, который выполняет проверку подписи.
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

// readRequestBody читает тело HTTP-запроса и сохраняет его в буфер.
//
// Функция используется для сохранения тела запроса для последующей обработки,
// так как тело запроса может быть прочитано только один раз.
//
// Параметры:
// - r: HTTP-запрос, тело которого нужно прочитать.
//
// Возвращаемые значения:
// - bytes.Buffer: Буфер, содержащий тело запроса.
// - error: Ошибка, если произошла проблема при чтении тела запроса.
func readRequestBody(r *http.Request) (bytes.Buffer, error) {
	var tempBuf bytes.Buffer
	teeReader := io.TeeReader(r.Body, &tempBuf)
	_, err := io.ReadAll(teeReader)
	return tempBuf, err
}

// validateSignature проверяет подпись запроса.
//
// Подпись считается действительной, если хеш, вычисленный на сервере,
// совпадает с хешем из заголовка "HashSHA256".
//
// Параметры:
// - body: Тело запроса, для которого вычисляется подпись.
// - agentSignStr: Подпись из заголовка "HashSHA256".
// - signKey: Ключ, используемый для вычисления подписи.
//
// Возвращаемое значение:
// - bool: true, если подпись действительна; false в противном случае.
func validateSignature(body []byte, agentSignStr string, signKey []byte) bool {
	serverSign := calculateServerSignature(body, signKey)

	agentSign, err := hex.DecodeString(agentSignStr)
	if err != nil || len(agentSign) != 32 {
		return false
	}

	return serverSign == [32]byte(agentSign)
}

// calculateServerSignature вычисляет подпись сервера для тела запроса.
//
// Подпись вычисляется как SHA-256 хеш от конкатенации тела запроса,
// запятой и ключа подписи.
//
// Параметры:
// - body: Тело запроса, для которого вычисляется подпись.
// - signKey: Ключ, используемый для вычисления подписи.
//
// Возвращаемое значение:
// - [32]byte: Хеш SHA-256, представляющий подпись сервера.
func calculateServerSignature(body []byte, signKey []byte) [32]byte {
	signBuf := append(body, ',')
	signBuf = append(signBuf, signKey...)
	return sha256.Sum256(signBuf)
}

// signData создает цифровую подпись данных с использованием ключа подписи.
func signData(data []byte, signKey []byte) string {
	if len(signKey) == 0 {
		return ""
	}

	signBuf := append(data, ',')
	signBuf = append(signBuf, signKey...)

	sign := sha256.Sum256(signBuf)

	return hex.EncodeToString(sign[:])
}

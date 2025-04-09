package logging

import (
	"net/http"
	"time"
)

// responseWriterWrapper — это обертка для ResponseWriter, которая записывает HTTP-статус
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// WriteHeader перехватывает вызов WriteHeader для записи статуса
func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}

	rw.wroteHeader = true
	rw.statusCode = statusCode

	rw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware логирование HTTP-запросов
func LoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			select {
			case <-ctx.Done():
				http.Error(w, "Request canceled", http.StatusServiceUnavailable)
				return
			default:
			}

			start := time.Now()

			// Извлечение username из контекста
			/*username, ok := r.Context().Value(UserContextKey).(string)
			if !ok {
				username = "unknown" // Если username отсутствует, используем "unknown"
			}

			// Чтение тела запроса
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				// Восстанавливаем тело запроса, чтобы оно могло быть использовано дальше
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			maskedBody := logging.MaskSensitiveData(string(bodyBytes))*/

			logger.Info("incoming request",
				//"username", username,
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				//"body", maskedBody, // Логируем тело запроса
			)

			rww := &responseWriterWrapper{ResponseWriter: w}

			next.ServeHTTP(rww, r)

			duration := time.Since(start)
			logger.Info("request completed",
				//	"username", username,
				"method", r.Method,
				"url", r.URL.String(),
				"status_code", rww.statusCode,
				"duration", duration.String(),
			)
		})
	}
}

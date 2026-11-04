package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging logs the start and end of every request with structured fields.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			id, err := NewRequestID()
			if err != nil {
				id = "unknown"
			}

			// Attach to context
			ctx := ContextWithRequestID(r.Context(), id)
			r = r.WithContext(ctx)

			// Set response header
			w.Header().Set("X-Request-ID", id)

			// Wrap response writer to capture status
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Log request start
			start := time.Now()
			logger.Info("request started",
				"request_id", id,
				"method", r.Method,
				"path", r.URL.Path,
			)

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log request completion
			duration := time.Since(start)
			logger.Info("request finished",
				"request_id", id,
				"status", rw.status,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

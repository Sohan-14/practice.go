package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery catches panics from downstream handlers and converts them
// into a 500 response, preventing the server from crashing.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := RequestIDFromContext(r.Context())

					// Log the full detail server-side
					logger.Error("recovered from panic",
						"request_id", requestID,
						"panic", rec,
						"stack", string(debug.Stack()),
					)

					// Return a clean JSON error to the client
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"internal_server_error","message":"An unexpected error occurred"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
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

type contextKey string

const requestIDKey contextKey = "request_id"

var logger *slog.Logger

func main() {

	// Initialize structured JSON logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/hello", helloHandler)

	handler := recoverMiddleware(loggingMiddleware(mux))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// serverErr carries fatal startup errors (e.g. port already in use)
	// from the server goroutine back to main.
	serverErr := make(chan error, 1)

	go func() {
		log.Println("server starting on :8080")
		// http.ErrServerClosed is the expected result of a graceful Shutdown,
		// so we only report real errors.
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for either a shutdown signal or a startup failure.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("server failed to start: %v", err)
	case sig := <-sigChan:
		log.Printf("received signal %v, shutting down", sig)
	}

	// Graceful drain, capped at 10 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("graceful shutdown complete")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(requestIDKey).(string)
	logger.Info("home handler executing", "request_id", id)

	time.Sleep(100 * time.Millisecond) // simulate work

	fmt.Fprintln(w, "Welcome to my Go HTTP server!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(requestIDKey).(string)
	logger.Info("health check: querying db", "request_id", id)

	time.Sleep(50 * time.Millisecond) // simulate work

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(requestIDKey).(string)
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	logger.Info("hello handler executing", "request_id", id, "name", name)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hello, %s!", name),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("recovered from panic: %v\n%s", rec, debug.Stack())

				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Internal Server Error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Generate request ID
		id, err := newRequestID()
		if err != nil {
			id = "unknown" // Fallback if random generation fails
		}

		// NEW: Return the ID to the client
		w.Header().Set("X-Request-ID", id)

		// 2. Attach to context
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		r = r.WithContext(ctx)

		// 3. Wrap the response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// 4. Log the start
		start := time.Now()
		logger.Info("request started",
			"request_id", id,
			"method", r.Method,
			"path", r.URL.Path,
		)

		// 5. Execute the actual handler
		next.ServeHTTP(rw, r)

		// 6. Log the completion
		duration := time.Since(start)
		logger.Info("request finished",
			"request_id", id,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

func newRequestID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

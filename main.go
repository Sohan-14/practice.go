package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/hello", helloHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
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
	fmt.Fprintln(w, "Welcome to my Go HTTP server!")
	fmt.Fprintln(w, "Try: /health or /api/hello?name=World")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hello, %s!", name),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

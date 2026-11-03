package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/hello", helloHandler)

	addr := ":8080"
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Server starting on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// homeHandler returns a welcome message
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to my Go HTTP server!")
	fmt.Fprintln(w, "Try: /health or /api/hello?name=World")
}

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// helloHandler returns a personalized greeting
func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hello, %s!", name),
	})
}

// writeJSON is a helper to send JSON responses
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

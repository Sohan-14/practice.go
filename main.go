package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
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

type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []validationError `json:"details,omitempty"`
}

// clientBucket holds the state for a single client's token bucket.
type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter manages token buckets for multiple clients.
type RateLimiter struct {
	mu         sync.Mutex // Protects the clients map
	clients    map[string]*clientBucket
	capacity   float64 // Max tokens (burst limit)
	refillRate float64 // Tokens added per second
}

// NewRateLimiter creates a rate limiter with the given capacity and refill rate.
func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		clients:    make(map[string]*clientBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// Allow checks if a request from the given clientID should be allowed.
// Returns (allowed, retryAfter) where retryAfter is how long to wait if not allowed.
func (rl *RateLimiter) Allow(clientID string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clients[clientID]
	now := time.Now()

	if !exists {
		bucket = &clientBucket{
			tokens:     rl.capacity,
			lastRefill: now,
		}
		rl.clients[clientID] = bucket
	} else {
		elapsed := now.Sub(bucket.lastRefill).Seconds()
		bucket.tokens += elapsed * rl.refillRate

		if bucket.tokens > rl.capacity {
			bucket.tokens = rl.capacity
		}

		bucket.lastRefill = now
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true, 0 // Allowed, no retry needed
	}

	// Not allowed: calculate how long until we have 1 token
	// We need (1.0 - current_tokens) more tokens
	// At refillRate tokens per second, that takes:
	tokensNeeded := 1.0 - bucket.tokens
	retryAfterSeconds := tokensNeeded / rl.refillRate
	retryAfter := time.Duration(retryAfterSeconds * float64(time.Second))

	return false, retryAfter
}

// Cleanup removes entries that haven't been used for longer than maxAge.
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, bucket := range rl.clients {
		if now.Sub(bucket.lastRefill) > maxAge {
			delete(rl.clients, clientID)
		}
	}
}

// StartCleanup launches a background goroutine that periodically cleans up stale entries.
func (rl *RateLimiter) StartCleanup(interval, maxAge time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			rl.Cleanup(maxAge)
		}
	}()
}

func main() {

	// Initialize structured JSON logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// NEW: Create a rate limiter
	// Capacity: 10 requests (burst), Refill rate: 2 tokens per second
	rateLimiter := NewRateLimiter(10.0, 2.0)

	// NEW: Start background cleanup
	// Run cleanup every 5 minutes, remove entries older than 1 hour
	rateLimiter.StartCleanup(5*time.Minute, 1*time.Hour)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/hello", helloHandler)
	mux.HandleFunc("POST /api/users", createUserHandler)
	mux.HandleFunc("GET /api/search", searchHandler)

	// Apply middleware layers (inside-out order)
	// 1. mux (routing)
	// 2. loggingMiddleware (logs request start/finish)
	// 3. rateLimitMiddleware (enforces rate limits)
	// 4. recoverMiddleware (catches panics)
	handler := recoverMiddleware(
		rateLimitMiddleware(rateLimiter)(
			loggingMiddleware(mux),
		),
	)

	httpsServer := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			// Force TLS 1.2 or higher (disable TLS 1.0 and 1.1)
			MinVersion: tls.VersionTLS12,

			// Use only strong cipher suites
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},

			// Prefer server's cipher order (not client's)
			PreferServerCipherSuites: true,
		},
	}

	// NEW: HTTP redirect server
	httpRedirectMux := http.NewServeMux()
	httpRedirectMux.HandleFunc("/", redirectHandler)

	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpRedirectMux,
	}

	// serverErr carries fatal startup errors (e.g. port already in use)
	// from the server goroutine back to main.
	serverErr := make(chan error, 2)

	// Start HTTP redirect server
	go func() {
		log.Println("HTTP redirect server starting on :80")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Start HTTPS server
	go func() {
		log.Println("HTTPS server starting on :8080")
		if err := httpsServer.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
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

	// Shutdown both servers
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}
	if err := httpsServer.Shutdown(ctx); err != nil {
		log.Printf("HTTPS server forced shutdown: %v", err)
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
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_json",
			Message: "Request body contains unknown fields",
		})
		return
	}

	// Collect all validation errors
	var errors []validationError

	if user.Name == "" {
		errors = append(errors, validationError{
			Field:   "name",
			Message: "Name is required",
		})
	} else if len(user.Name) < 2 {
		errors = append(errors, validationError{
			Field:   "name",
			Message: "Name must be at least 2 characters",
		})
	}

	if user.Email == "" {
		errors = append(errors, validationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else if !isValidEmail(user.Email) {
		errors = append(errors, validationError{
			Field:   "email",
			Message: "Email format is invalid",
		})
	}

	// If there are validation errors, return them
	if len(errors) > 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed",
			Details: errors,
		})
		return
	}

	// Success case
	log.Printf("created user: %s (%s)", user.Name, user.Email)
	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "User created",
		"name":    user.Name,
		"email":   user.Email,
	})
}
func isValidEmail(email string) bool {
	return strings.Contains(email, "@")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Collect validation errors
	var errors []validationError

	// Required parameter: q
	q := query.Get("q")
	if q == "" {
		errors = append(errors, validationError{
			Field:   "q",
			Message: "Search query is required",
		})
	}

	// Optional parameter: page (default 1, must be positive)
	pageStr := query.Get("page")
	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			errors = append(errors, validationError{
				Field:   "page",
				Message: "Page must be a number",
			})
		} else if p < 1 {
			errors = append(errors, validationError{
				Field:   "page",
				Message: "Page must be at least 1",
			})
		} else {
			page = p
		}
	}

	// Optional parameter: limit (default 10, must be 1-100)
	limitStr := query.Get("limit")
	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			errors = append(errors, validationError{
				Field:   "limit",
				Message: "Limit must be a number",
			})
		} else if l < 1 || l > 100 {
			errors = append(errors, validationError{
				Field:   "limit",
				Message: "Limit must be between 1 and 100",
			})
		} else {
			limit = l
		}
	}

	// Return validation errors if any
	if len(errors) > 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "validation_failed",
			Message: "Query parameter validation failed",
			Details: errors,
		})
		return
	}

	// Success: return the parsed parameters
	writeJSON(w, http.StatusOK, map[string]any{
		"query": q,
		"page":  page,
		"limit": limit,
	})
}
func rateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			allowed, retryAfter := rl.Allow(clientIP)

			if !allowed {
				// Set the Retry-After header (in seconds, rounded up)
				retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))

				// Return structured error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error":       "rate_limit_exceeded",
					"message":     "Too many requests. Please try again later.",
					"retry_after": strconv.Itoa(retryAfterSeconds),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// RemoteAddr is in the format "IP:port"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, just return RemoteAddr as-is
		return r.RemoteAddr
	}
	return ip
}

// redirectHandler redirects all HTTP requests to HTTPS.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Build the HTTPS URL
	httpsURL := "https://" + r.Host + r.URL.RequestURI()

	// 301 = permanent redirect (browsers cache this)
	// 302 = temporary redirect (browsers don't cache)
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}
